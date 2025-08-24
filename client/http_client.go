package client

import (
	"bytes"
	"goload/metrics"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	client *http.Client
}

type RequestOptions struct {
	Headers map[string]string
	Cookies []*http.Cookie
	Body    []byte // store body as []byte instead of io.Reader
}

// DoRequest performs any HTTP method with headers, cookies, and body
func (c *HttpClient) DoRequest(method, url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	startTime := time.Now()

	var bodyReader io.Reader
	var sentBytes int64
	if opts != nil && opts.Body != nil {
		sentBytes = int64(len(opts.Body))
		bodyReader = bytes.NewReader(opts.Body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	// Add headers
	if opts != nil && opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	// Add cookies
	if opts != nil && opts.Cookies != nil {
		for _, cookie := range opts.Cookies {
			req.AddCookie(cookie)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Read all response bytes
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	requestMetric := &metrics.RequestMetric{
		Duration:   time.Since(startTime),
		StatusCode: resp.StatusCode,
	}

	networkMetric := &metrics.NetworkMetric{
		BytesSent: sentBytes,
		BytesRecv: int64(len(responseBytes)),
	}

	return requestMetric, networkMetric, nil
}

// Convenience methods
func (c *HttpClient) Get(url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	return c.DoRequest(http.MethodGet, url, opts)
}

func (c *HttpClient) Post(url string, contentType string, body []byte, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	opts.Headers["Content-Type"] = contentType
	opts.Body = body
	return c.DoRequest(http.MethodPost, url, opts)
}

func (c *HttpClient) Put(url string, contentType string, body []byte, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}
	if opts.Headers == nil {
		opts.Headers = make(map[string]string)
	}
	opts.Headers["Content-Type"] = contentType
	opts.Body = body
	return c.DoRequest(http.MethodPut, url, opts)
}

func (c *HttpClient) Delete(url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	return c.DoRequest(http.MethodDelete, url, opts)
}
