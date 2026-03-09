package twitterinternalapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	graphqlBaseURL   = "https://x.com/i/api/graphql"
	graphqlReferrer  = "https://x.com/"
	graphqlFetchSite = "same-origin"
)

// Client represents a Twitter API client authenticated with a token.
type Client struct {
	authToken  string
	csrfToken  string
	httpClient *http.Client
	cookies    string
	tidKey     string
	tidGen     *TransactionIDGenerator
	Tweets     *TweetsService
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

// GraphQLRequest represents a GraphQL request with variables and features
type GraphQLRequest struct {
	Variables map[string]interface{} `json:"variables"`
	QueryID   string                 `json:"queryId"`
	Features  map[string]bool        `json:"features"`
}

// ExecuteGraphQL executes a GraphQL query with variables, queryID, operation name, and features
func (c *Client) ExecuteGraphQL(
	variables map[string]interface{},
	queryID string,
	operationName string,
	features map[string]bool,
) (map[string]interface{}, error) {
	// Construct the GraphQL request
	payload := GraphQLRequest{
		Variables: variables,
		QueryID:   queryID,
		Features:  features,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL - format: https://x.com/i/api/graphql/{queryId}/{operationName}
	finalURL := fmt.Sprintf("%s/%s/%s", graphqlBaseURL, queryID, operationName)

	req, err := http.NewRequest("POST", finalURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("X-Twitter-Active-User", "yes")
	req.Header.Set("X-Twitter-Auth-Type", "OAuth2Session")
	req.Header.Set("X-Twitter-Client-Language", "en")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", graphqlFetchSite)
	req.Header.Set("Sec-CH-UA", `"Not(A:Brand";v="8", "Chromium";v="144"`)
	req.Header.Set("Sec-CH-UA-Mobile", "?0")
	req.Header.Set("Sec-CH-UA-Platform", `"macOS"`)
	req.Header.Set("Referer", graphqlReferrer)

	if c.tidGen != nil && c.tidKey != "" {
		transactionID := c.tidGen.GenerateHeader(finalURL, "POST", c.tidKey)
		if transactionID != "" {
			req.Header.Set("X-Client-Transaction-ID", transactionID)
		}
	}

	if c.csrfToken != "" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
		req.AddCookie(&http.Cookie{
			Name:  "ct0",
			Value: c.csrfToken,
		})
	}

	if c.authToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: c.authToken,
		})
	}

	if c.cookies != "" {
		for _, cookie := range strings.Split(c.cookies, ";") {
			log.Printf("Adding cookie: %s\n", cookie)
			parts := strings.SplitN(cookie, "=", 2)
			if len(parts) == 2 {
				req.AddCookie(&http.Cookie{
					Name:  strings.TrimSpace(parts[0]),
					Value: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(result["message"].(string)))
	}

	if errs, ok := result["errors"].([]interface{}); ok && len(errs) > 0 {
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

	return result, nil
}
