package goload

import (
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
	Headers    []http.Header `yaml:"headers,omitempty"`
	Cookies    []http.Cookie `yaml:"cookies,omitempty"`
	Error      error
}

type Request struct {
	Method    HttpMethod    `yaml:"method"`
	URI       string        `yaml:"uri"`
	UserAgent UserAgent     `yaml:"user_agent"`
	Headers   http.Header   `yaml:"headers"`
	Body      string        `yaml:"body"`
	Cookies   []http.Cookie `yaml:"cookies"`
}
type Collection struct {
	Name  string `yaml:"name"`
	Tests []Test `yaml:"tests"`
}

type Test struct {
	Name    *string `yaml:"name"`
	Global  *Global `yaml:"global,omitempty"`
	Request Request `yaml:"request"`
	Phases  []Phase `yaml:"phases"`
}
type Global struct {
	Timeout      time.Duration `yaml:"timeout,omitempty"` // e.g., "30s"
	Retries      int           `yaml:"retries,omitempty"` // Number of retries per request
	RetriesDelay time.Duration `yaml:"retries_delay,omitempty"`
	ThinkTime    time.Duration `yaml:"think_time,omitempty"` // Delay between requests per VU
}
type Phase struct {
	Name          *string        `yaml:"name"`
	SingleRequest *bool          `yaml:"single_request,omitempty"`
	Duration      *time.Duration `yaml:"duration,omitempty"`
	StartVUs      *int           `yaml:"start_vus,omitempty"`
	EndVUs        *int           `yaml:"end_vus,omitempty"`
	TargetVUs     *int           `yaml:"target_vus,omitempty"`
}

type LogConfig struct {
	LogOutputPath  string `yaml:"logOutputPath"`
	FilenameFormat string `yaml:"filenameFormat"`
}
