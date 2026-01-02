// Package api provides the HTTP client for KEF speaker API.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client communicates with the KEF speaker HTTP API.
type Client struct {
	host       string
	port       int
	httpClient *http.Client
	ctx        context.Context
}

// NewClient creates a new API client.
func NewClient(host string, port int, timeout time.Duration) *Client {
	return &Client{
		host: host,
		port: port,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		ctx: context.Background(),
	}
}

// SetHost updates the target host.
func (c *Client) SetHost(host string) {
	c.host = host
}

// SetContext sets the context for requests.
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// GetData performs a GET request to /api/getData.
func (c *Client) GetData(path, roles string) ([]interface{}, error) {
	if c.host == "" {
		return nil, fmt.Errorf("no host configured")
	}

	params := url.Values{}
	params.Set("path", path)
	params.Set("roles", roles)

	apiURL := fmt.Sprintf("http://%s:%d/api/getData?%s", c.host, c.port, params.Encode())
	req, err := http.NewRequestWithContext(c.ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// SetData performs a GET request to /api/setData.
func (c *Client) SetData(path, roles, value string) error {
	if c.host == "" {
		return fmt.Errorf("no host configured")
	}

	params := url.Values{}
	params.Set("path", path)
	params.Set("roles", roles)
	params.Set("value", value)

	apiURL := fmt.Sprintf("http://%s:%d/api/setData?%s", c.host, c.port, params.Encode())
	req, err := http.NewRequestWithContext(c.ctx, "GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return nil
}

// GetInt retrieves an integer value from the API.
func (c *Client) GetInt(path string) (int, error) {
	result, err := c.GetData(path, "value")
	if err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, fmt.Errorf("empty response")
	}

	data, ok := result[0].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	v, ok := data["i32_"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid integer format")
	}

	return int(v), nil
}

// GetString retrieves a string value from the API.
func (c *Client) GetString(path string) (string, error) {
	result, err := c.GetData(path, "value")
	if err != nil {
		return "", err
	}

	if len(result) == 0 {
		return "", fmt.Errorf("empty response")
	}

	data, ok := result[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	v, ok := data["string_"].(string)
	if !ok {
		return "", fmt.Errorf("invalid string format")
	}

	return v, nil
}

// SetInt sets an integer value via the API.
func (c *Client) SetInt(path string, value int) error {
	jsonValue := fmt.Sprintf(`{"type":"i32_","i32_":%d}`, value)
	return c.SetData(path, "value", jsonValue)
}
