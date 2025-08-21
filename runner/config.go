package goload

import (
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

type Response struct {
	StatusCode int
	Body       string
	Headers    []http.Header
	Cookies    []http.Cookie
}

type Request struct {
	Method    HttpMethod    `yaml:"method"`
	URI       string        `yaml:"URI"`
	UserAgent UserAgent     `yaml:"userAgent"`
	Headers   http.Header   `yaml:"headers"`
	Body      string        `yaml:"body"`
	Cookies   []http.Cookie `yaml:"cookies"`
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

type Config struct {
	Test Test `yaml:"test"`
}

type Test struct {
	Name   string  `yaml:"name"`
	Global Global  `yaml:"global"`
	Phases []Phase `yaml:"phases"`
}
type Global struct {
	Timeout   time.Duration `yaml:"timeout"`    // e.g., "30s"
	Retries   int           `yaml:"retries"`    // Number of retries per request
	ThinkTime time.Duration `yaml:"think_time"` // Delay between requests per VU
}
type Phase struct {
	Name          string        `yaml:"name"`
	Duration      time.Duration `yaml:"duration,omitempty"`
	StartVUs      *int          `yaml:"start_vus,omitempty"`
	EndVUs        *int          `yaml:"end_vus,omitempty"`
	TargetVUs     *int          `yaml:"target_vus,omitempty"`
	ReqsPerSecond *int          `yaml:"reqs_per_second,omitempty"`
	Mode          string        `yaml:"mode,omitempty"`
	Phases        []NestedPhase `yaml:"phases,omitempty"` // nested phases
}
type NestedPhase struct {
	Phase     string        `yaml:"phase"`
	Duration  time.Duration `yaml:"duration"`
	TargetVUs int           `yaml:"target_vus"`
}

type LogConfig struct {
	LogOutputPath  string `yaml:"logOutputPath"`
	FilenameFormat string `yaml:"filenameFormat"`
}
