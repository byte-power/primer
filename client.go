package primer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	SandboxBaseURL    = "https://api.sandbox.primer.io"
	ProductionBaseURL = "https://api.primer.io"

	apiVersion = "2.4"
)

var defaultHTTPClient = &http.Client{
	Timeout: 90 * time.Second,
}

// RequestOption configures optional per-request behaviour (e.g. idempotency).
type RequestOption func(h http.Header)

// WithIdempotencyKey attaches an X-Idempotency-Key header to the request.
// Primer uses this to deduplicate retried mutations so the same side-effect
// (charge, refund, …) is applied at most once.
// See https://primer.io/docs/api-reference/get-started/idempotency-key
func WithIdempotencyKey(key string) RequestOption {
	return func(h http.Header) {
		h.Set("X-Idempotency-Key", key)
	}
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	logger     Logger
}

// NewClient 创建 Primer 客户端（默认使用 Sandbox 环境）
func NewClient(apiKey string, logger Logger) *Client {
	return NewClientWithBaseURL(apiKey, SandboxBaseURL, logger)
}

// NewProductionClient 创建 Primer 生产环境客户端
func NewProductionClient(apiKey string, logger Logger) *Client {
	return NewClientWithBaseURL(apiKey, ProductionBaseURL, logger)
}

// NewClientWithBaseURL 使用自定义 Base URL 创建 Primer 客户端
func NewClientWithBaseURL(apiKey, baseURL string, logger Logger) *Client {
	if logger == nil {
		logger = &NopLogger{}
	}
	return &Client{
		httpClient: defaultHTTPClient,
		baseURL:    baseURL,
		apiKey:     apiKey,
		logger:     logger,
	}
}

// NewClientWithHTTPClient 使用自定义 HTTP 客户端创建 Primer 客户端
func NewClientWithHTTPClient(apiKey, baseURL string, httpClient *http.Client, logger Logger) *Client {
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}
	if logger == nil {
		logger = &NopLogger{}
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		apiKey:     apiKey,
		logger:     logger,
	}
}

// doGet 执行 HTTP GET 请求
func (c *Client) doGet(endpoint string, params url.Values, response any) *Error {
	reqURL := c.baseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return WrapError(err, "failed to create request")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("X-Api-Version", apiVersion)

	c.logger.Debug("primer_request",
		String("method", "GET"),
		String("url", reqURL),
		Any("headers", req.Header))

	return c.executeRequest(req, reqURL, response)
}

// doPost 执行 HTTP POST 请求
func (c *Client) doPost(endpoint string, body any, response any, opts ...RequestOption) *Error {
	return c.doRequest(http.MethodPost, endpoint, body, response, opts...)
}

// doPatch 执行 HTTP PATCH 请求
func (c *Client) doPatch(endpoint string, body any, response any, opts ...RequestOption) *Error {
	return c.doRequest(http.MethodPatch, endpoint, body, response, opts...)
}

// doDelete 执行 HTTP DELETE 请求
func (c *Client) doDelete(endpoint string, response any, opts ...RequestOption) *Error {
	return c.doRequest(http.MethodDelete, endpoint, nil, response, opts...)
}

// doRequest 执行 HTTP 请求（POST/PATCH/PUT/DELETE）
func (c *Client) doRequest(method, endpoint string, body any, response any, opts ...RequestOption) *Error {
	reqURL := c.baseURL + endpoint

	var bodyReader io.Reader
	var bodyStr string
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return WrapError(err, "failed to marshal request body")
		}
		bodyReader = bytes.NewReader(data)
		bodyStr = string(data)
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return WrapError(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("X-Api-Version", apiVersion)

	for _, opt := range opts {
		opt(req.Header)
	}

	if bodyStr != "" {
		c.logger.Debug("primer_request",
			String("method", method),
			String("url", reqURL),
			Any("headers", req.Header),
			String("body", bodyStr))
	} else {
		c.logger.Debug("primer_request",
			String("method", method),
			String("url", reqURL),
			Any("headers", req.Header))
	}

	return c.executeRequest(req, reqURL, response)
}

// executeRequest 执行 HTTP 请求并解析响应
func (c *Client) executeRequest(req *http.Request, reqURL string, response any) *Error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("primer_request_error",
			String("method", req.Method),
			String("url", reqURL),
			ErrorField(err))
		return WrapError(err, "request failed")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("primer_read_response_error",
			String("url", reqURL),
			ErrorField(err))
		return WrapError(err, "failed to read response")
	}

	c.logger.Debug("primer_response",
		String("url", reqURL),
		Int("status_code", resp.StatusCode),
		Any("headers", resp.Header),
		String("body", string(respBody)))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseErrorResponse(resp.StatusCode, respBody, reqURL)
	}

	if response != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, response); err != nil {
			return WrapError(err, "failed to unmarshal response")
		}
	}

	return nil
}

// parseErrorResponse 解析 API 错误响应
func (c *Client) parseErrorResponse(statusCode int, body []byte, reqURL string) *Error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.ErrorID != "" {
		c.logger.Error("primer_api_error",
			String("url", reqURL),
			Int("status_code", statusCode),
			String("error_id", errResp.Error.ErrorID),
			String("description", errResp.Error.Description))
		return &Error{
			Message:    fmt.Sprintf("Primer API error: %s", errResp.Error.Description),
			StatusCode: statusCode,
			ErrorID:    errResp.Error.ErrorID,
		}
	}

	return &Error{
		Message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
		StatusCode: statusCode,
	}
}
