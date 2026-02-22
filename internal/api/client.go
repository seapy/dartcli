package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/seapy/dartcli/internal/httpclient"
)

const baseURL = "https://opendart.fss.or.kr"

// Client is a DART OpenAPI HTTP client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New creates a new API client with the given API key.
func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: httpclient.New(30 * time.Second),
	}
}

// APIError represents a non-OK DART API status.
type APIError struct {
	Status  string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("DART API error %s: %s", e.Status, e.Message)
}

// get performs a GET request and decodes JSON into dst.
func (c *Client) get(path string, params url.Values, dst interface{}) error {
	params.Set("crtfc_key", c.apiKey)
	u := fmt.Sprintf("%s%s?%s", baseURL, path, params.Encode())

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}

	return nil
}

// getRaw performs a GET request and returns the raw bytes.
func (c *Client) getRaw(path string, params url.Values) ([]byte, error) {
	params.Set("crtfc_key", c.apiKey)
	u := fmt.Sprintf("%s%s?%s", baseURL, path, params.Encode())

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// checkStatus inspects the BaseResponse and returns an APIError if status != "000".
func checkStatus(base BaseResponse) error {
	if base.Status != "000" {
		return &APIError{Status: base.Status, Message: base.Message}
	}
	return nil
}
