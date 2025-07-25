package goload

import (
	"net/http"
	"time"
)

type ExecutionMode int
type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	DELETE HttpMethod = "DELETE"
	HEAD   HttpMethod = "HEAD"
	PATCH  HttpMethod = "PATCH"
)

type ExecutionTimepoint struct {
	Duration time.Duration
	TargetVu int
}

type Response struct {
	StatusCode int
	Body       string
	Headers    []http.Header
	Cookies    []http.Cookie
}

type Config struct {
	Request       Request
	Response      Response
	Log           bool
	LogOutputPath string
	Duration      time.Duration
	Timeout       time.Duration
	Timepoints    []ExecutionTimepoint
}

type Request struct {
	Method    HttpMethod
	URI       string
	UserAgent UserAgent
	Headers   http.Header
	Body      string
	Cookies   []http.Cookie
}

type UserAgent string

const (
	ChromeAgent  UserAgent = "Chrome"
	FirefoxAgent UserAgent = "Firefox"
	SafariAgent  UserAgent = "Safari"
	EdgeAgent    UserAgent = "Edge"
	OperaAgent   UserAgent = "Opera"
	IEAgent      UserAgent = "IE"
	AndroidAgent UserAgent = "Android"
	IOSAgent     UserAgent = "IOS"
)
