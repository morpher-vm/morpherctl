package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		timeout  time.Duration
		token    string
		expected *Client
	}{
		{
			name:    "should create client with default timeout",
			baseURL: "http://localhost:9000",
			timeout: 0,
			token:   "test-token",
			expected: &Client{
				baseURL: "http://localhost:9000",
				timeout: 30 * time.Second,
				token:   "test-token",
			},
		},
		{
			name:    "should create client with custom timeout",
			baseURL: "http://localhost:9000",
			timeout: 60 * time.Second,
			token:   "test-token",
			expected: &Client{
				baseURL: "http://localhost:9000",
				timeout: 60 * time.Second,
				token:   "test-token",
			},
		},
		{
			name:    "should create client without token",
			baseURL: "http://localhost:9000",
			timeout: 30 * time.Second,
			token:   "",
			expected: &Client{
				baseURL: "http://localhost:9000",
				timeout: 30 * time.Second,
				token:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.timeout, tt.token)

			assert.Equal(t, tt.expected.baseURL, client.GetBaseURL())
			assert.Equal(t, tt.expected.timeout, client.GetTimeout())
			assert.Equal(t, tt.expected.token, client.token)
			assert.NotNil(t, client.httpClient)
		})
	}
}

func TestClient_Ping(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		responseTime    string
		expectedSuccess bool
		withToken       bool
	}{
		{
			name:            "should return success for 200 response",
			statusCode:      200,
			responseTime:    "10ms",
			expectedSuccess: true,
			withToken:       false,
		},
		{
			name:            "should return failure for 500 response",
			statusCode:      500,
			responseTime:    "",
			expectedSuccess: false,
			withToken:       false,
		},
		{
			name:            "should work with authorization token",
			statusCode:      200,
			responseTime:    "15ms",
			expectedSuccess: true,
			withToken:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if authorization header is set when token is provided.
				if tt.withToken {
					authHeader := r.Header.Get("Authorization")
					assert.Equal(t, "Bearer test-token", authHeader)
				}

				// Set response time header if provided.
				if tt.responseTime != "" {
					w.Header().Set("X-Response-Time", tt.responseTime)
				}

				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte("OK"))
				require.NoError(t, err)
			}))
			defer server.Close()

			// Create client.
			token := ""
			if tt.withToken {
				token = "test-token"
			}
			client := NewClient(server.URL, 30*time.Second, token)

			// Test ping.
			ctx := context.Background()
			response, err := client.Ping(ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.statusCode, response.StatusCode)
			assert.Equal(t, tt.responseTime, response.ResponseTime)
			assert.Equal(t, tt.expectedSuccess, response.Success)
		})
	}
}

// testHTTPResponse is a helper function to test HTTP responses.
func testHTTPResponse(t *testing.T, path string, statusCode int, expectedSuccess bool) {
	// Create test server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, path, r.URL.Path)
		w.WriteHeader(statusCode)

		// Return appropriate response based on path.
		switch path {
		case "/info":
			if statusCode == 200 {
				// Return valid JSON for info endpoint.
				jsonResponse := `{
					"OS": {
						"Name": "darwin",
						"PlatformName": "darwin",
						"PlatformVersion": "15.3.1",
						"KernelVersion": "24.3.0"
					},
					"GoVersion": "go1.25.0",
					"UpTime": "32s"
				}`
				_, err := w.Write([]byte(jsonResponse))
				require.NoError(t, err)
			} else {
				_, err := w.Write([]byte("Error"))
				require.NoError(t, err)
			}
		default:
			_, err := w.Write([]byte("OK"))
			require.NoError(t, err)
		}
	}))
	defer server.Close()

	// Create client.
	client := NewClient(server.URL, 30*time.Second, "")

	// Test the request.
	ctx := context.Background()
	var response any
	var err error

	switch path {
	case "/info":
		response, err = client.GetInfo(ctx)
	default:
		t.Fatalf("unknown path: %s", path)
	}

	require.NoError(t, err)

	switch r := response.(type) {
	case *InfoResponse:
		assert.Equal(t, statusCode, r.StatusCode)
		assert.Equal(t, expectedSuccess, r.Success)
	default:
		t.Fatalf("unexpected response type: %T", response)
	}
}

func TestClient_GetInfo(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		expectedSuccess bool
	}{
		{
			name:            "should return success for 200 response",
			statusCode:      200,
			expectedSuccess: true,
		},
		{
			name:            "should return failure for 500 response",
			statusCode:      500,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHTTPResponse(t, "/info", tt.statusCode, tt.expectedSuccess)
		})
	}
}

func TestClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		expectedHealthy bool
	}{
		{
			name:            "should return healthy for 200 response",
			statusCode:      200,
			expectedHealthy: true,
		},
		{
			name:            "should return unhealthy for 500 response",
			statusCode:      500,
			expectedHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte("OK"))
				require.NoError(t, err)
			}))
			defer server.Close()

			// Create client.
			client := NewClient(server.URL, 30*time.Second, "")

			// Test health check.
			ctx := context.Background()
			healthy, err := client.IsHealthy(ctx)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedHealthy, healthy)
		})
	}
}

func TestClient_AuthorizationHeader(t *testing.T) {
	// Create test server that checks authorization header.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer test-token" {
			w.WriteHeader(200)
			_, err := w.Write([]byte("Authorized"))
			require.NoError(t, err)
		} else {
			w.WriteHeader(401)
			_, err := w.Write([]byte("Unauthorized"))
			require.NoError(t, err)
		}
	}))
	defer server.Close()

	t.Run("should include authorization header when token is provided", func(t *testing.T) {
		client := NewClient(server.URL, 30*time.Second, "test-token")

		ctx := context.Background()
		response, err := client.Ping(ctx)
		require.NoError(t, err)

		assert.Equal(t, 200, response.StatusCode)
		assert.True(t, response.Success)
	})

	t.Run("should not include authorization header when token is empty", func(t *testing.T) {
		client := NewClient(server.URL, 30*time.Second, "")

		ctx := context.Background()
		response, err := client.Ping(ctx)
		require.NoError(t, err)

		assert.Equal(t, 401, response.StatusCode)
		assert.False(t, response.Success)
	})
}
