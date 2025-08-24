package client

import (
	"goload/metrics"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	client *http.Client
}

// RequestOptions allows custom headers, cookies, and body
type RequestOptions struct {
	Headers map[string]string
	Cookies []*http.Cookie
	Body    io.Reader
}

// DoRequest performs any HTTP method with headers, cookies, and body
func (c *HttpClient) doRequest(method, url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	startTime := time.Now()

	var bodyReader io.Reader
	var sentBytes int64
	if opts != nil && opts.Body != nil {
		// Wrap body to count bytes sent
		bodyReader = &countingReader{R: opts.Body, Count: &sentBytes}
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

	// Count response bytes
	var receivedBytes int64
	respBody := &countingReader{R: resp.Body, Count: &receivedBytes}
	_, err = io.ReadAll(respBody)
	if err != nil {
		return nil, nil, err
	}

	requestMetric := &metrics.RequestMetric{
		Duration:   time.Since(startTime),
		StatusCode: resp.StatusCode,
	}
	networkMetric := &metrics.NetworkMetric{
		BytesSent: sentBytes,
		BytesRecv: receivedBytes,
	}

	return requestMetric, networkMetric, nil
}

type countingReader struct {
	R     io.Reader
	Count *int64
}

func (c *countingReader) Read(p []byte) (n int, err error) {
	n, err = c.R.Read(p)
	*c.Count += int64(n)
	return
}

func (c *HttpClient) Get(url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	return c.doRequest(http.MethodGet, url, opts)
}

func (c *HttpClient) Post(url string, contentType string, body io.Reader, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}
	opts.Headers = map[string]string{"Content-Type": contentType}
	opts.Body = body
	return c.doRequest(http.MethodPost, url, opts)
}

func (c *HttpClient) Put(url string, contentType string, body io.Reader, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}
	opts.Headers = map[string]string{"Content-Type": contentType}
	opts.Body = body
	return c.doRequest(http.MethodPut, url, opts)
}

func (c *HttpClient) Delete(url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	return c.doRequest(http.MethodDelete, url, opts)
}
