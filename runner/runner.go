package main

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
	duration   time.Duration
	initialVus int
	timepoints map[string]int
}

type Config struct {
	Method        HttpMethod
	URI           string
	Vus           int
	Timeout       time.Duration
	Log           bool
	LogOutputPath string
	Execution     Execution
}
