package client

import (
	"goload/metrics"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	DELETE HttpMethod = "DELETE"
	HEAD   HttpMethod = "HEAD"
	PATCH  HttpMethod = "PATCH"
)

type UserAgent string

const (
	ChromeAgent  UserAgent = "chrome"
	FirefoxAgent UserAgent = "firefox"
	SafariAgent  UserAgent = "safari"
	EdgeAgent    UserAgent = "edge"
	OperaAgent   UserAgent = "opera"
	IEAgent      UserAgent = "ie"
	AndroidAgent UserAgent = "android"
	IOSAgent     UserAgent = "ios"
)

type HTTPResponse struct {
	StatusCode    int `yaml:"status_code,omitempty"`
	Body          string
	Headers       []HTTPClientHeader `yaml:"headers,omitempty"`
	Cookies       []HTTPClientCookie `yaml:"cookies,omitempty"`
	RequestMetric *metrics.RequestMetric
	NetworkMetric *metrics.NetworkMetric
	Error         error
}

type HTTPRequest struct {
	Method    HttpMethod         `yaml:"method"`
	URI       string             `yaml:"uri"`
	UserAgent UserAgent          `yaml:"user_agent"`
	Headers   []HTTPClientHeader `yaml:"headers"`
	Body      string             `yaml:"body"`
	Cookies   []HTTPClientCookie `yaml:"cookies"`
}

type HTTPClientHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

func (requestHeader *HTTPClientHeader) convertToHeader() http.Header {
	header := http.Header{}
	header.Set(requestHeader.Name, requestHeader.Value)
	return header
}

func (requestHeader *HTTPClientHeader) parse(header http.Header) HTTPClientHeader {
	return HTTPClientHeader{
		Name:  requestHeader.Name,
		Value: header.Get(requestHeader.Name),
	}
}

type HTTPClientCookie struct {
	Name       string    `yaml:"name"`
	Value      string    `yaml:"value"`
	Path       string    `yaml:"path"`
	Domain     string    `yaml:"domain"`
	Secure     bool      `yaml:"secure"`
	HTTPOnly   bool      `yaml:"http_only"`
	Expires    time.Time `yaml:"expires"`
	MaxAge     int       `yaml:"max_age"`
	SameSite   string    `yaml:"same_site"`
	RawExpires string    `yaml:"raw_expires"`
	RawMaxAge  string    `yaml:"raw_max_age"`
}

func (requestCookie *HTTPClientCookie) convertToCookie() http.Cookie {
	return http.Cookie{
		Name:       requestCookie.Name,
		Value:      requestCookie.Value,
		Path:       requestCookie.Path,
		Domain:     requestCookie.Domain,
		Expires:    requestCookie.Expires,
		RawExpires: requestCookie.RawExpires,
		MaxAge:     requestCookie.MaxAge,
	}
}

func (requestCookie *HTTPClientCookie) parse(cookie http.Cookie) HTTPClientCookie {
	return HTTPClientCookie{
		Name:       cookie.Name,
		Value:      cookie.Value,
		Path:       cookie.Path,
		Domain:     cookie.Domain,
		Expires:    cookie.Expires,
		RawExpires: cookie.RawExpires,
		MaxAge:     cookie.MaxAge,
	}
}

type Client struct {
	HttpClient *http.Client
}

type RequestOptions struct {
	Headers []HTTPClientHeader
	Cookies []HTTPClientCookie
	Body    *string
}

func (c *Client) ExecuteRequest(req *http.Request) (*HTTPResponse, error) {
	startTime := time.Now()
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return &HTTPResponse{
			Error: err,
			RequestMetric: &metrics.RequestMetric{
				Duration: time.Since(startTime),
			},
		}, err
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{
			StatusCode: resp.StatusCode,
			Error:      err,
			RequestMetric: &metrics.RequestMetric{
				Duration:   time.Since(startTime),
				StatusCode: resp.StatusCode,
			},
		}, err
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Convert headers
	var headers []HTTPClientHeader
	for name, values := range resp.Header {
		for _, value := range values {
			headers = append(headers, HTTPClientHeader{
				Name:  name,
				Value: value,
			})
		}
	}

	// Convert cookies
	var cookies []HTTPClientCookie
	for _, cookie := range resp.Cookies() {
		cookies = append(cookies, HTTPClientCookie{
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

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       string(bodyBytes),
		Headers:    headers,
		Cookies:    cookies,
		RequestMetric: &metrics.RequestMetric{
			Duration:   duration,
			StatusCode: resp.StatusCode,
		},
		NetworkMetric: &metrics.NetworkMetric{
			BytesSent: 0,
			BytesRecv: int64(len(bodyBytes)),
		},
		Error: nil,
	}, nil
}

func CreateRequest(request HTTPRequest) (*http.Request, error) {
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
