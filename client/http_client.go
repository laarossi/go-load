package client

import (
	"bytes"
	"goload/metrics"
	"io"
	"net/http"
	"time"
)

type HttpMethod string

const (
	GET    HttpMethod = "get"
	POST   HttpMethod = "post"
	PUT    HttpMethod = "put"
	DELETE HttpMethod = "delete"
	HEAD   HttpMethod = "head"
	PATCH  HttpMethod = "patch"
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

type Response struct {
	StatusCode int `yaml:"status_code,omitempty"`
	Body       string
	Duration   time.Duration
	Headers    []RequestHeader `yaml:"headers,omitempty"`
	Cookies    []RequestCookie `yaml:"cookies,omitempty"`
	Error      error
}

type Request struct {
	Method    HttpMethod      `yaml:"method"`
	URI       string          `yaml:"uri"`
	UserAgent UserAgent       `yaml:"user_agent"`
	Headers   []RequestHeader `yaml:"headers"`
	Body      string          `yaml:"body"`
	Cookies   []RequestCookie `yaml:"cookies"`
}

type RequestHeader struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type RequestCookie struct {
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

type Client struct {
	HttpClient *http.Client
}

func parseSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

type RequestOptions struct {
	Headers []RequestHeader
	Cookies []RequestCookie
	Body    *string
}

// DoRequest performs any HTTP method with headers, cookies, and body
func (c *Client) DoRequest(method, url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	startTime := time.Now()

	var bodyReader io.Reader
	var sentBytes int64
	if opts != nil && opts.Body != nil {
		sentBytes = int64(len(*opts.Body))
		bodyReader = bytes.NewReader([]byte(*opts.Body))
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	// Add headers
	if opts != nil && opts.Headers != nil {
		for _, v := range opts.Headers {
			req.Header.Set(v.Name, v.Value)
		}
	}

	// Add cookies
	if opts != nil && opts.Cookies != nil {
		for _, cookie := range opts.Cookies {
			req.AddCookie(&http.Cookie{
				Name:       cookie.Name,
				Value:      cookie.Value,
				Path:       cookie.Path,
				Domain:     cookie.Domain,
				Expires:    cookie.Expires,
				RawExpires: cookie.RawExpires,
				MaxAge:     cookie.MaxAge,
				Secure:     cookie.Secure,
				HttpOnly:   cookie.HTTPOnly,
				SameSite:   parseSameSite(cookie.SameSite),
				Raw:        "",
				Unparsed:   nil,
			})
		}
	}

	resp, err := c.HttpClient.Do(req)
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
func (c *Client) Get(url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	return c.DoRequest(http.MethodGet, url, opts)
}

func (c *Client) Post(url string, contentType string, body []byte, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}
	if opts.Headers == nil {
		opts.Headers = make([]RequestHeader, 0)
	}

	opts.Headers = append(opts.Headers, RequestHeader{Name: "Content-Type", Value: contentType})
	s := string(body)
	opts.Body = &s
	return c.DoRequest(http.MethodPost, url, opts)
}

func (c *Client) Put(url string, contentType string, body []byte, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}
	if opts.Headers == nil {
		opts.Headers = make([]RequestHeader, 0)
	}
	opts.Headers = append(opts.Headers, RequestHeader{Name: "Content-Type", Value: contentType})
	s := string(body)
	opts.Body = &s
	return c.DoRequest(http.MethodPut, url, opts)
}

func (c *Client) Delete(url string, opts *RequestOptions) (*metrics.RequestMetric, *metrics.NetworkMetric, error) {
	return c.DoRequest(http.MethodDelete, url, opts)
}
