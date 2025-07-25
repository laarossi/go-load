package goload

import (
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

type Config struct {
	Method        HttpMethod
	URI           string
	UserAgent     string
	Log           bool
	LogOutputPath string
	Duration      time.Duration
	Timeout       time.Duration
	Timepoints    []ExecutionTimepoint
}
