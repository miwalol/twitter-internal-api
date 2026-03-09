package twitterinternalapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	graphqlBaseURL   = "https://x.com/i/api/graphql"
	graphqlReferrer  = "https://x.com/"
	graphqlFetchSite = "same-origin"
	bearerToken      = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
)

// Client represents a Twitter API client authenticated with a token.
type Client struct {
	authToken       string
	csrfToken       string
	mu              sync.RWMutex
	httpClient      *http.Client
	cookies         string
	tidKey          string
	tidGen          *TransactionIDGenerator
	Tweets          *TweetsService
	onCSRFRefreshed func(newToken string)
}

// NewClient creates a new Twitter API client with the given auth token.
func NewClient(authToken, csrfToken string) *Client {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	c := &Client{
		authToken:  authToken,
		csrfToken:  csrfToken,
		httpClient: httpClient,
	}

	c.Tweets = &TweetsService{client: c}
	return c
}

// SetCSRFToken sets the CSRF token for requests
func (c *Client) SetCSRFToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.csrfToken = token
}

// SetCookies sets the cookies for requests
func (c *Client) SetCookies(cookies string) {
	c.cookies = cookies
}

// SetTransactionIDGenerator sets the transaction ID generator with frames and key
func (c *Client) SetTransactionIDGenerator(key string, frames [][][]int) {
	c.tidKey = key
	c.tidGen = NewTransactionIDGenerator(frames)
}

// SetHTTPClient allows you to provide a custom HTTP client.
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// GetCSRFToken returns the current CSRF token (ct0 cookie value)
func (c *Client) GetCSRFToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.csrfToken
}

// OnCSRFRefreshed registers a callback invoked whenever the ct0 token is refreshed from a response.
// Useful for persisting the updated token to a cache or storage.
func (c *Client) OnCSRFRefreshed(fn func(newToken string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onCSRFRefreshed = fn
}

// setCSRFTokenLocked sets the CSRF token in a thread-safe manner (internal use)
func (c *Client) setCSRFTokenLocked(token string) {
	c.mu.Lock()
	c.csrfToken = token
	fn := c.onCSRFRefreshed
	c.mu.Unlock()
	if fn != nil {
		fn(token)
	}
}

// extractCT0FromCookie extracts the ct0 value from a Set-Cookie header
func extractCT0FromCookie(cookieHeader string) string {
	parts := strings.Split(cookieHeader, ";")
	if len(parts) == 0 {
		return ""
	}

	mainPart := strings.TrimSpace(parts[0])
	keyValue := strings.SplitN(mainPart, "=", 2)
	if len(keyValue) != 2 {
		return ""
	}

	cookieName := strings.TrimSpace(keyValue[0])
	if cookieName == "ct0" {
		return strings.TrimSpace(keyValue[1])
	}

	return ""
}

// prepareRequest attaches the static bearer token, CSRF token, and auth cookies to req.
func (c *Client) prepareRequest(req *http.Request) {
	req.Header.Set("Authorization", bearerToken)
	csrfToken := c.GetCSRFToken()
	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
		req.AddCookie(&http.Cookie{Name: "ct0", Value: csrfToken})
	}
	if c.authToken != "" {
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: c.authToken})
	}
}

// applyCommonHeaders sets browser-like headers and extra cookies shared across all requests.
func (c *Client) applyCommonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", graphqlFetchSite)
	req.Header.Set("Sec-CH-UA", `"Not(A:Brand";v="8", "Chromium";v="144"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
	req.Header.Set("Referer", graphqlReferrer)

	if c.cookies != "" {
		for _, cookie := range strings.Split(c.cookies, ";") {
			parts := strings.SplitN(cookie, "=", 2)
			if len(parts) == 2 {
				req.AddCookie(&http.Cookie{
					Name:  strings.TrimSpace(parts[0]),
					Value: strings.TrimSpace(parts[1]),
				})
			}
		}
	}
}

// applyClientHeaders sets headers that are only applicable to GraphQL/client requests.
func (c *Client) applyClientHeaders(req *http.Request) {
	req.Header.Set("X-Twitter-Active-User", "yes")
	req.Header.Set("X-Twitter-Client-Language", "en")

	if c.tidGen != nil && c.tidKey != "" {
		transactionID := c.tidGen.GenerateHeader(req.URL.String(), req.Method, c.tidKey)
		if transactionID != "" {
			req.Header.Set("X-Client-Transaction-ID", transactionID)
		}
	}
}

// GraphQLRequest represents a GraphQL request with variables and features
type GraphQLRequest struct {
	Variables map[string]interface{} `json:"variables"`
	QueryID   string                 `json:"queryId"`
	Features  map[string]bool        `json:"features"`
}

// executeGraphQLRequest performs the HTTP request for a GraphQL call and handles the response.
// It is used by ExecuteGraphQL and for retries on CSRF expiry (error code 353).
// retry controls whether a single retry is attempted on error code 353.
func (c *Client) executeGraphQLRequest(body []byte, queryID, operationName string) (map[string]interface{}, error) {
	return c.doGraphQLRequest(body, queryID, operationName, true)
}

func (c *Client) doGraphQLRequest(body []byte, queryID, operationName string, retry bool) (map[string]interface{}, error) {
	finalURL := fmt.Sprintf("%s/%s/%s", graphqlBaseURL, queryID, operationName)

	req, err := http.NewRequest("POST", finalURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.prepareRequest(req)
	c.applyCommonHeaders(req)
	c.applyClientHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	for _, setCookie := range resp.Header["Set-Cookie"] {
		if ct0 := extractCT0FromCookie(setCookie); ct0 != "" {
			c.setCSRFTokenLocked(ct0)
			break
		}
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if errs, ok := result["errors"].([]interface{}); ok && len(errs) > 0 {
		if retry {
			for _, e := range errs {
				if errMap, ok := e.(map[string]interface{}); ok {
					if code, ok := errMap["code"].(float64); ok && int(code) == 353 {
						// CSRF token expired — ct0 may have been refreshed via Set-Cookie above; retry once
						return c.doGraphQLRequest(body, queryID, operationName, false)
					}
				}
			}
		}
		errMessages := make([]string, len(errs))
		for i, e := range errs {
			if errMap, ok := e.(map[string]interface{}); ok {
				if msg, ok := errMap["message"].(string); ok {
					errMessages[i] = msg
				}
			}
		}
		return nil, fmt.Errorf("GraphQL error: %s", strings.Join(errMessages, ", "))
	}

	if resp.StatusCode >= 400 {
		if message, ok := result["message"].(string); ok {
			return nil, errors.New(message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return result, nil
}

// ExecuteGraphQL executes a GraphQL query with variables, queryID, operation name, and features
func (c *Client) ExecuteGraphQL(
	variables map[string]interface{},
	queryID string,
	operationName string,
	features map[string]bool,
) (map[string]interface{}, error) {
	payload := GraphQLRequest{
		Variables: variables,
		QueryID:   queryID,
		Features:  features,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.executeGraphQLRequest(body, queryID, operationName)
}
