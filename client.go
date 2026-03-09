package twitterinternalapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Twitter API client authenticated with a token.
type Client struct {
	authToken  string
	httpClient *http.Client
	baseURL    string
	Tweets     *TweetsService
}

// NewClient creates a new Twitter API client with the given auth token.
func NewClient(authToken string) *Client {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	c := &Client{
		authToken:  authToken,
		httpClient: httpClient,
		baseURL:    "https://api.twitter.com/graphql",
	}

	c.Tweets = &TweetsService{client: c}
	return c
}

// SetHTTPClient allows you to provide a custom HTTP client.
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// executeGraphQL executes a GraphQL query against the Twitter API.
func (c *Client) executeGraphQL(query string, variables map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	req.Header.Set("User-Agent", "TwitterAndroid/10.31.0 (312310000)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if errs, ok := result["errors"]; ok {
		return nil, fmt.Errorf("GraphQL error: %v", errs)
	}

	return result, nil
}
