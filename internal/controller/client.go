package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"morpherctl/internal/config"
)

const (
	defaultControllerURL = "http://localhost:9000"
	defaultTimeout       = "30s"
)

// Client handles communication with the morpher controller.
type Client struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
	token      string
}

// PingResponse represents the response from a ping request.
type PingResponse struct {
	StatusCode   int    `json:"status_code"`
	ResponseTime string `json:"response_time,omitempty"`
	Success      bool   `json:"success"`
}

// OSInfo represents operating system information.
type OSInfo struct {
	Name            string `json:"Name"`
	PlatformName    string `json:"PlatformName"`
	PlatformVersion string `json:"PlatformVersion"`
	KernelVersion   string `json:"KernelVersion"`
}

// InfoResult represents the result data in info response.
type InfoResult struct {
	OS        OSInfo `json:"OS"`
	GoVersion string `json:"GoVersion"`
	UpTime    string `json:"UpTime"`
}

// InfoResponse represents the response from an info request.
type InfoResponse struct {
	StatusCode int         `json:"status_code"`
	Success    bool        `json:"success"`
	Result     *InfoResult `json:"result,omitempty"`
}

// NewClient creates a new controller client.
func NewClient(baseURL string, timeout time.Duration, token string) *Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: baseURL,
		timeout: timeout,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		token: token,
	}
}

// GetControllerConfig retrieves common configuration values for controller commands.
func GetControllerConfig() (string, time.Duration, string, error) {
	// Get configuration values.
	configMgr := config.NewManager("")

	controllerURL, err := configMgr.GetString("controller.url")
	if err != nil {
		controllerURL = defaultControllerURL
	}

	timeoutStr, err := configMgr.GetString("controller.timeout")
	if err != nil {
		timeoutStr = defaultTimeout
	}

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 30 * time.Second
	}

	token, err := configMgr.GetString("auth.token")
	if err != nil {
		token = ""
	}

	return controllerURL, timeout, token, nil
}

// CreateControllerClient creates a new controller client with configuration.
func CreateControllerClient() (*Client, time.Duration, error) {
	controllerURL, timeout, token, err := GetControllerConfig()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get controller configuration: %w", err)
	}

	client := NewClient(controllerURL, timeout, token)
	return client, timeout, nil
}

// newRequest creates a new HTTP request with context and authorization.
func (c *Client) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

// Ping sends a ping request to the controller.
func (c *Client) Ping(ctx context.Context) (*PingResponse, error) {
	req, err := c.newRequest(ctx, "GET", "/ping")
	if err != nil {
		return nil, fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send ping request: %w", err)
	}
	defer resp.Body.Close()

	response := &PingResponse{
		StatusCode:   resp.StatusCode,
		ResponseTime: resp.Header.Get("X-Response-Time"),
		Success:      resp.StatusCode == http.StatusOK,
	}

	return response, nil
}

// GetInfo retrieves detailed controller information.
func (c *Client) GetInfo(ctx context.Context) (*InfoResponse, error) {
	req, err := c.newRequest(ctx, "GET", "/info")
	if err != nil {
		return nil, fmt.Errorf("failed to create info request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller info: %w", err)
	}
	defer resp.Body.Close()

	response := &InfoResponse{
		StatusCode: resp.StatusCode,
		Success:    resp.StatusCode == http.StatusOK,
	}

	// If the request was successful, try to parse the response body.
	if resp.StatusCode == http.StatusOK {
		var result InfoResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return response, fmt.Errorf("failed to parse controller info response: %w", err)
		}
		response.Result = &result
	}

	return response, nil
}

// IsHealthy checks if the controller is healthy based on ping response.
func (c *Client) IsHealthy(ctx context.Context) (bool, error) {
	response, err := c.Ping(ctx)
	if err != nil {
		return false, err
	}
	return response.Success, nil
}

// GetBaseURL returns the base URL of the controller.
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// GetTimeout returns the timeout setting of the client.
func (c *Client) GetTimeout() time.Duration {
	return c.timeout
}
