package types

import (
	"fmt"
	"net/http"
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
	RequestMetric *RequestMetric
	NetworkMetric *NetworkMetric
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

func (r *HTTPRequest) Summary() string {
	return fmt.Sprintf("Method: %s | URI: %s | UserAgent: %s | Body: %s | Headers size: %d | Cookies size: %d", r.Method, r.URI, r.UserAgent, r.Body, len(r.Headers), len(r.Cookies))
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
