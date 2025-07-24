package main

import (
	goload "goload/runner"
	"time"
)

func main() {
	config := goload.Config{
		Method:        goload.GET,
		URI:           "https://www.google.com",
		UserAgent:     "PostmanRuntime/7.44.1",
		LogOutputPath: "logs",
		Vus:           2,
		Execution: goload.Execution{
			Duration: 2 * time.Second,
		},
	}

	duration, err := time.ParseDuration("30s")
	if err != nil {
		panic(err)
	}

	config.Timeout = duration
	goload.Execute(config)
}
