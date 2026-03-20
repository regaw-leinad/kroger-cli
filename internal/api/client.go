package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/regaw-leinad/kroger-cli/internal/auth"
	"github.com/regaw-leinad/kroger-cli/internal/secrets"
)

const baseURL = "https://api.kroger.com/v1"

// Client is a Kroger API client that handles authentication and token refresh.
type Client struct {
	http         *http.Client
	store        *secrets.Store
	clientID     string
	clientSecret string
}

// NewClient creates a new API client.
func NewClient(store *secrets.Store, clientID, clientSecret string) *Client {
	return &Client{
		http:         &http.Client{Timeout: 30 * time.Second},
		store:        store,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// Get performs an authenticated GET request and decodes the JSON response into v.
func (c *Client) Get(path string, params url.Values, v any) error {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if err := c.setAuth(req); err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// Put performs an authenticated PUT request with a JSON body.
func (c *Client) Put(path string, body any, v any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest("PUT", baseURL+path, strings.NewReader(string(jsonBody)))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if err := c.setAuth(req); err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// setAuth adds the Bearer token to the request, refreshing if needed.
func (c *Client) setAuth(req *http.Request) error {
	tok, err := c.store.GetToken()
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}
	if tok.AccessToken == "" {
		return fmt.Errorf("not logged in - run 'kroger auth login' first")
	}

	if tok.IsExpired() {
		tok, err = auth.RefreshToken(c.clientID, c.clientSecret, tok.RefreshToken)
		if err != nil {
			return fmt.Errorf("refreshing token: %w (try 'kroger auth login')", err)
		}
		if err := c.store.SetToken(tok); err != nil {
			return fmt.Errorf("storing refreshed token: %w", err)
		}
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	return nil
}

// APIError represents a non-2xx response from the Kroger API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	reason := parseErrorReason(e.Body)
	if reason != "" {
		return fmt.Sprintf("%s (HTTP %d)", reason, e.StatusCode)
	}
	return fmt.Sprintf("API error (HTTP %d)", e.StatusCode)
}

func parseErrorReason(body string) string {
	// Kroger returns {"errors":{"reason":"Service Unavailable",...}}
	var parsed struct {
		Errors struct {
			Reason string `json:"reason"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err == nil && parsed.Errors.Reason != "" {
		return parsed.Errors.Reason
	}
	return ""
}
