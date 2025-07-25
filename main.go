package main

import (
	goload "goload/runner"
	"time"
)

func main() {
	config := goload.Config{
		Method:        goload.GET,
		URI:           "http://localhost:8974/products/100",
		UserAgent:     "PostmanRuntime/7.44.1",
		LogOutputPath: "logs",
		Timepoints: []goload.ExecutionTimepoint{
			{
				Duration: time.Second * 10,
				TargetVu: 10,
			},
			{
				Duration: time.Second * 10,
				TargetVu: 5,
			},
			{
				Duration: time.Second * 15,
				TargetVu: 2,
			},
		},
	}

	duration, err := time.ParseDuration("30s")
	if err != nil {
		panic(err)
	}

	config.Timeout = duration
	goload.Execute(config)
}
