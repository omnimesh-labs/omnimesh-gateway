package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient is a test HTTP client with helper methods
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient creates a new test HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// JSONRequest represents a JSON request
type JSONRequest struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// JSONResponse represents a JSON response
type JSONResponse struct {
	StatusCode int
	Body       map[string]interface{}
	Headers    http.Header
	Raw        []byte
}

// DoJSON performs a JSON request and returns a JSON response
func (c *HTTPClient) DoJSON(req JSONRequest) (*JSONResponse, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	url := c.baseURL + req.Path
	httpReq, err := http.NewRequest(req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &JSONResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Raw:        rawBody,
	}

	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &response.Body); err != nil {
			// If JSON parsing fails, just keep the raw body
			response.Body = map[string]interface{}{
				"_raw": string(rawBody),
			}
		}
	}

	return response, nil
}

// Get performs a GET request
func (c *HTTPClient) Get(path string, headers ...map[string]string) (*JSONResponse, error) {
	req := JSONRequest{
		Method: "GET",
		Path:   path,
	}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return c.DoJSON(req)
}

// Post performs a POST request with JSON body
func (c *HTTPClient) Post(path string, body interface{}, headers ...map[string]string) (*JSONResponse, error) {
	req := JSONRequest{
		Method: "POST",
		Path:   path,
		Body:   body,
	}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return c.DoJSON(req)
}

// Put performs a PUT request with JSON body
func (c *HTTPClient) Put(path string, body interface{}, headers ...map[string]string) (*JSONResponse, error) {
	req := JSONRequest{
		Method: "PUT",
		Path:   path,
		Body:   body,
	}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return c.DoJSON(req)
}

// Delete performs a DELETE request
func (c *HTTPClient) Delete(path string, headers ...map[string]string) (*JSONResponse, error) {
	req := JSONRequest{
		Method: "DELETE",
		Path:   path,
	}
	if len(headers) > 0 {
		req.Headers = headers[0]
	}
	return c.DoJSON(req)
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      string      `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
	ID      string                 `json:"id"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DoJSONRPC performs a JSON-RPC request
func (c *HTTPClient) DoJSONRPC(path, method string, params interface{}, id string) (*JSONRPCResponse, *JSONResponse, error) {
	if id == "" {
		id = fmt.Sprintf("test-%d", time.Now().UnixNano())
	}

	rpcReq := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      id,
	}

	resp, err := c.Post(path, rpcReq)
	if err != nil {
		return nil, resp, err
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(resp.Raw, &rpcResp); err != nil {
		return nil, resp, fmt.Errorf("failed to parse JSON-RPC response: %w", err)
	}

	return &rpcResp, resp, nil
}
