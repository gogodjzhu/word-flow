package util

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// HTTPClientBuilder provides a fluent interface for building HTTP requests
type HTTPClientBuilder struct {
	client  *http.Client
	request *http.Request
}

// NewHTTPClientBuilder creates a new HTTP client builder
func NewHTTPClientBuilder() *HTTPClientBuilder {
	return &HTTPClientBuilder{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// WithTimeout sets custom timeout
func (b *HTTPClientBuilder) WithTimeout(timeout time.Duration) *HTTPClientBuilder {
	b.client.Timeout = timeout
	return b
}

// Get creates a GET request
func (b *HTTPClientBuilder) Get(url string, headers map[string]string) *HTTPClientBuilder {
	return b.buildRequest("GET", url, nil, nil, headers)
}

// Post creates a POST request with JSON content
func (b *HTTPClientBuilder) Post(url string, body []byte, headers map[string]string) *HTTPClientBuilder {
	return b.buildRequest("POST", url, body, map[string]string{
		"Content-Type": "application/json",
		"Accept":       "*/*",
	}, headers)
}

// PostStream creates a POST request for streaming
func (b *HTTPClientBuilder) PostStream(url string, body []byte, headers map[string]string) *HTTPClientBuilder {
	return b.buildRequest("POST", url, body, map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "text/event-stream",
		"Cache-Control": "no-cache",
	}, headers)
}

// buildRequest builds the HTTP request with common headers
func (b *HTTPClientBuilder) buildRequest(method, url string, body []byte, defaultHeaders, customHeaders map[string]string) *HTTPClientBuilder {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return b
	}

	// Set default headers
	for k, v := range defaultHeaders {
		req.Header.Set(k, v)
	}

	// Set content length for POST requests
	if body != nil {
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}

	// Set host header
	if host, err := hostname(url); err == nil {
		req.Header.Set("Host", host)
	}

	// Set custom headers
	for k, v := range customHeaders {
		req.Header.Set(k, v)
	}

	// Set default user agent for GET requests
	if method == "GET" {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	}

	b.request = req
	return b
}

// Execute executes the request and processes response
func (b *HTTPClientBuilder) Execute(processor func(*http.Response) (interface{}, error)) (interface{}, error) {
	if b.request == nil {
		return nil, nil
	}

	resp, err := b.client.Do(b.request)
	if err != nil {
		return nil, err
	}

	return processor(resp)
}

// ExecuteStream executes streaming request
func (b *HTTPClientBuilder) ExecuteStream(processor func(*http.Response) error) error {
	if b.request == nil {
		return nil
	}

	resp, err := b.client.Do(b.request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return processor(resp)
}

// SendGet sends GET request (backward compatibility)
func SendGet(url string, header map[string]string, wrap func(response *http.Response) (interface{}, error)) (interface{}, error) {
	return NewHTTPClientBuilder().
		Get(url, header).
		Execute(wrap)
}

// SendPost sends POST request (backward compatibility)
func SendPost(url string, header map[string]string, body []byte, wrap func(response *http.Response) (interface{}, error)) (interface{}, error) {
	return NewHTTPClientBuilder().
		Post(url, body, header).
		Execute(wrap)
}

// SendPostStream sends streaming POST request (backward compatibility)
func SendPostStream(url string, header map[string]string, body []byte, wrap func(response *http.Response) error) error {
	return NewHTTPClientBuilder().
		PostStream(url, body, header).
		ExecuteStream(wrap)
}

// SendPostWithTimeout sends POST request with custom timeout (backward compatibility)
func SendPostWithTimeout(url string, header map[string]string, body []byte, timeout time.Duration, wrap func(response *http.Response) (interface{}, error)) (interface{}, error) {
	return NewHTTPClientBuilder().
		WithTimeout(timeout).
		Post(url, body, header).
		Execute(wrap)
}

// SendPostStreamWithTimeout sends streaming POST request with custom timeout (backward compatibility)
func SendPostStreamWithTimeout(url string, header map[string]string, body []byte, timeout time.Duration, wrap func(response *http.Response) error) error {
	return NewHTTPClientBuilder().
		WithTimeout(timeout).
		PostStream(url, body, header).
		ExecuteStream(wrap)
}

func hostname(u string) (string, error) {
	e, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return e.Hostname(), nil
}
