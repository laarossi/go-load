package client

import (
	"goload/types"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	HttpClient *http.Client
}

type RequestOptions struct {
	Headers []types.HTTPClientHeader
	Cookies []types.HTTPClientCookie
	Body    *string
}

func (c *Client) ExecuteRequest(req *http.Request) (*types.HTTPResponse, error) {
	startTime := time.Now()
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return &types.HTTPResponse{
			Error: err,
			RequestMetric: &types.RequestMetric{
				Duration: time.Since(startTime),
			},
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &types.HTTPResponse{
			StatusCode: resp.StatusCode,
			Error:      err,
			RequestMetric: &types.RequestMetric{
				Duration:   time.Since(startTime),
				StatusCode: resp.StatusCode,
			},
		}, err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Convert headers
	var headers []types.HTTPClientHeader
	for name, values := range resp.Header {
		for _, value := range values {
			headers = append(headers, types.HTTPClientHeader{
				Name:  name,
				Value: value,
			})
		}
	}

	// Convert cookies
	var cookies []types.HTTPClientCookie
	for _, cookie := range resp.Cookies() {
		cookies = append(cookies, types.HTTPClientCookie{
			Name:       cookie.Name,
			Value:      cookie.Value,
			Path:       cookie.Path,
			Domain:     cookie.Domain,
			Expires:    cookie.Expires,
			RawExpires: cookie.RawExpires,
			MaxAge:     cookie.MaxAge,
			Secure:     cookie.Secure,
			HTTPOnly:   cookie.HttpOnly,
		})
	}

	return &types.HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
		Headers:    headers,
		Cookies:    cookies,
		RequestMetric: &types.RequestMetric{
			Duration:   duration,
			StatusCode: resp.StatusCode,
		},
		NetworkMetric: &types.NetworkMetric{
			BytesSent: 0,
			BytesRecv: int64(len(bodyBytes)),
		},
		Error: nil,
	}, nil
}

func CreateRequest(request types.HTTPRequest) (*http.Request, error) {
	parsedURL, err := url.Parse(request.URI)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	for _, header := range request.Headers {
		headers.Add(header.Name, header.Value)
	}

	req := &http.Request{
		Method: "GET",
		URL:    parsedURL,
		Body:   io.NopCloser(strings.NewReader(request.Body)),
		Header: headers,
		Host:   parsedURL.Host,
		Proto:  "HTTP/1.1",
	}

	return req, nil
}
