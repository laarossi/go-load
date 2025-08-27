# GoLoad

GoLoad is a flexible and powerful HTTP load-testing library for Go. It helps you stress test and benchmark web services with configurable virtual users, schedules, and realistic user-agent simulation.

GoLoad works by either providing a yaml configuration file or programmatically by defining a collection of tests and their phases.
- Minimum Go version: 1.22+

## Features

- Supports YAML test configuration
- Multiple HTTP methods: GET, POST, PUT, DELETE, HEAD, PATCH
- User-agent simulation (Chrome, Firefox, Safari, Edge, Opera, IE, Android, iOS)
- Full control over requests: headers, cookies, and bodies
- Test-wide duration and per-request timeout settings
- Request and response logging
- Flexible execution using phases

## Quick Start

**goload** provides 2 type of runtime execution :
* YAML configuration file
* Programmatic configuration

**Programmatic :**
```textmate
package main

import (
	"goload/internal/runner"
	"goload/types"
	"net/http"
)

func main() {
	collection := runner.Collection{
		Name: "collection-name",
	}
	req := types.HTTPRequest{
		Body:      "hello world",
		URI:       "http://localhost:8080/hello",
		UserAgent: types.ChromeAgent,
		Method:    http.MethodGet,
		Headers: []types.HTTPClientHeader{
			{Name: "Content-Type", Value: "application/json"},
			{Name: "Accept", Value: "application/json"},
		},
		Cookies: []types.HTTPClientCookie{
			{Name: "cookie1", Value: "value1"},
			{Name: "cookie2", Value: "value2"},
		},
	}
	phases := []runner.Phase{
		{
			Name:          "phase-1",
			SingleRequest: true,
		},
		{
			Name:         "phase-2",
			Duration:     "2s",
			TargetVUs:    5,
			Increment:    "200ms",
			IncrementVus: 10,
		},
	}
	tests := []runner.Test{
		{
			Name:    "test-name-1",
			Request: req,
			Phases:  phases,
		},
	}
	collection.Tests = tests
	executor, _ := runner.NewExecutor(collection)
	executor.Collection = collection
	executor.Execute()
}

```
**YAML :**

````text
name: Test collection
tests:
  # SPIKE TESTING
  - name: Spike testing
    thresholds:
      pass_if:
        - metric: latency_ms.p95
          target: "<=200"
        - metric: error_rate_pct
          target: "<=0.1"
      fail_if:
        - metric: availability
          target: "<90%"
    global:
      timeout: 30s
      retries: 5
      retries_delay: 2s
      think_time: 200ms
    request:
      method: get
      uri: http://localhost:8974/products/171
      user_agent: chrome
    response:
      status_code: 200
    phases:
      - duration: 5s
        target_vus: 5
        increment: 1s
        increment_vus: 5

````

````textmate
package main

import (
	"fmt"
	"goload/internal/runner"
)

func main() {
	executor, err := runner.LoadFromYaml("config/spike.yml")
	if err != nil {
		fmt.Printf("Failed to load config: %s\n", err)
	}
	executor.Execute()
}
````

## Usage

### Defining requests

```textmate
req := types.HTTPRequest{
			Body:      "hello world",
			URI:       "http://localhost:8080/hello",
			UserAgent: types.ChromeAgent,
			Method:    http.MethodGet,
			Headers: []types.HTTPClientHeader{
				{Name: "Content-Type", Value: "application/json"},
				{Name: "Accept", Value: "application/json"},
			},
			Cookies: []types.HTTPClientCookie{
				{Name: "cookie1", Value: "value1"},
				{Name: "cookie2", Value: "value2"},
			},
		}
```


- Method: GET, POST, PUT, DELETE, HEAD, PATCH are supported.
- Headers/Cookies: Provide custom values.
- Body: Raw bytes for JSON, form data, etc.

To replace the main request in each phase, you must provide a request to the phase, like the example below:

```textmate
func main() {
	collection := runner.Collection{..../}
	mainReq := types.HTTPRequest{..../}
	phaseReq := types.HTTPRequest{..../}
	phases := []runner.Phase{
		{
			..../
		},
		{
			Name:         "phase-2",
			Duration:     "2s",
			TargetVUs:    5,
			Increment:    "200ms",
			IncrementVus: 10,
			Request: &phaseReq
		},
	}
	..../
	executor.Execute()
}
```

### Execution phases



```textmate
type Phase struct {
	Name          string             `yaml:"name"`
	Duration      string             `yaml:"duration,omitempty"`
	Increment     string             `yaml:"increment,omitempty"`
	IncrementVus  int                `yaml:"increment_vus,omitempty"`
	TargetVUs     int                `yaml:"target_vus,omitempty"`
	Request       *types.HTTPRequest `yaml:"request"`
    SingleRequest bool               `yaml:"single_request,omitempty"`
}
```

- **SingleRequest** : **bool** value used to define if the test consists of only 1 request. When provided, only the **Name** and the **Request** parameters are allowed.
- **Duration** : the whole duration of the phase execution.
- **TargetVUs** : the initial number of vus when starting the execution
- **IncrementVus** : the number of virtual users added to target vus on each incrementation
- **Increment** : represents the duration of each increment
- **Request** : represents the request object, if provided it will override the global request provided
