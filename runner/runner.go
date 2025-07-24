package goload

import "time"

type ExecutionMode int
type HttpMethod string

const (
	GET    HttpMethod = "get"
	POST   HttpMethod = "post"
	PUT    HttpMethod = "put"
	DELETE HttpMethod = "delete"
	HEAD   HttpMethod = "head"
	PATCH  HttpMethod = "patch"
)

type Execution struct {
	Duration   time.Duration
	InitialVUs int
	Timepoints map[string]int
}

type Config struct {
	Method        HttpMethod
	URI           string
	UserAgent     string
	Vus           int
	Timeout       time.Duration
	Log           bool
	LogOutputPath string
	Execution     Execution
}
