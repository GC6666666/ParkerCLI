package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/parker/ParkerCli/pkg/logger"
)

// 常量定义
const (
	DefaultTimeout     = 30 * time.Second
	DefaultContentType = "application/json"
)

// HTTPClient 封装HTTP客户端功能
type HTTPClient struct {
	client     *http.Client
	baseURL    string
	headers    map[string]string
	retryCount int
	retryDelay time.Duration
}

// ClientOption 客户端选项函数类型
type ClientOption func(*HTTPClient)

// NewClient 创建HTTP客户端
func NewClient(options ...ClientOption) *HTTPClient {
	// 创建默认客户端
	c := &HTTPClient{
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		headers: map[string]string{
			"Content-Type": DefaultContentType,
			"User-Agent":   "ParkerCli-HTTPClient/0.1.0",
		},
		retryCount: 0,
		retryDelay: 1 * time.Second,
	}

	// 应用选项
	for _, option := range options {
		option(c)
	}

	return c
}

// WithTimeout 设置超时选项
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *HTTPClient) {
		c.client.Timeout = timeout
	}
}

// WithBaseURL 设置基础URL选项
func WithBaseURL(baseURL string) ClientOption {
	return func(c *HTTPClient) {
		c.baseURL = baseURL
	}
}

// WithHeader 设置HTTP头选项
func WithHeader(key, value string) ClientOption {
	return func(c *HTTPClient) {
		c.headers[key] = value
	}
}

// WithRetry 设置重试策略选项
func WithRetry(count int, delay time.Duration) ClientOption {
	return func(c *HTTPClient) {
		c.retryCount = count
		c.retryDelay = delay
	}
}

// Request 表示HTTP请求
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
	Query   map[string]string
}

// Response 表示HTTP响应
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// buildURL 构建完整URL
func (c *HTTPClient) buildURL(path string) string {
	if c.baseURL == "" {
		return path
	}

	baseURL := strings.TrimSuffix(c.baseURL, "/")
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("%s/%s", baseURL, path)
}

// Do 执行HTTP请求
func (c *HTTPClient) Do(ctx context.Context, req Request) (*Response, error) {
	var (
		resp    *http.Response
		err     error
		attempt int
	)

	// 构建URL
	url := c.buildURL(req.Path)

	// 准备请求体
	var reqBody io.Reader
	if req.Body != nil {
		jsonData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("请求体序列化失败: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 添加默认头
	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	// 添加请求特定头
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// 添加查询参数
	q := httpReq.URL.Query()
	for k, v := range req.Query {
		q.Add(k, v)
	}
	httpReq.URL.RawQuery = q.Encode()

	// 重试逻辑
	for attempt = 0; attempt <= c.retryCount; attempt++ {
		if attempt > 0 {
			logger.Info("HTTP请求重试 (%d/%d): %s %s", attempt, c.retryCount, req.Method, url)
			time.Sleep(c.retryDelay)
		}

		// 发送请求
		resp, err = c.client.Do(httpReq)
		if err == nil {
			break
		}

		// 检查上下文是否取消
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败 (重试 %d 次): %w", c.retryCount, err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 构建响应对象
	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// Get 执行GET请求
func (c *HTTPClient) Get(ctx context.Context, path string, query map[string]string) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
	})
}

// Post 执行POST请求
func (c *HTTPClient) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
}

// Put 执行PUT请求
func (c *HTTPClient) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodPut,
		Path:   path,
		Body:   body,
	})
}

// Delete 执行DELETE请求
func (c *HTTPClient) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, Request{
		Method: http.MethodDelete,
		Path:   path,
	})
}

// UnmarshalJSON 解析JSON响应到指定结构
func (r *Response) UnmarshalJSON(v interface{}) error {
	if len(r.Body) == 0 {
		return nil
	}
	return json.Unmarshal(r.Body, v)
}

// String 返回响应体字符串
func (r *Response) String() string {
	return string(r.Body)
}
