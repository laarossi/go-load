package runner

import (
	"encoding/json"
	"fmt"
	"goload/logger"
	"goload/metrics"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

type Executor struct {
	collection      Collection
	logger          logger.Logger
	metricCollector metrics.MetricsCollector
}

func LoadFromYaml(yamlFilePath string) (*Executor, error) {
	configPath := filepath.Join(yamlFilePath)
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return &Executor{}, err
	}

	var collection Collection
	err = yaml.Unmarshal(yamlFile, &collection)
	if err != nil {
		fmt.Printf("Error parsing YAML: %s\n", err)
		return &Executor{}, err
	}

	executor := Executor{
		collection: collection,
	}
	err = executor.load()
	if err != nil {
		return &Executor{}, err
	}
	return &executor, nil
}

func (e *Executor) load() error {
	e.logger, _ = logger.NewLogger("logs")
	e.metricCollector = metrics.MetricsCollector{}
	err := e.metricCollector.Init()
	if err != nil {
		return fmt.Errorf("error initializing metrics collector: %s", err)
	}
	return nil
}

func (e *Executor) Execute() {
	for _, test := range e.collection.Tests {
		if test.Name != nil {
			fmt.Println("parsing test configuration for ", *test.Name)
		} else {
			fmt.Println("parsing test configuration")
		}
		for _, phase := range test.Phases {
			err := e.executePhase(phase, test.Request, test.Global)
			if err != nil {
				_ = fmt.Errorf("failed to execute phase: %s", err)
			}
		}
	}
}

func (e *Executor) executePhase(phase Phase, request Request, global *Global) error {
	executionSegment, err := ResolvePhase(phase)
	if err != nil {
		fmt.Printf("Error resolving phase: %s\n", err)
		return nil
	}
	for {
		if executionSegment == nil {
			break
		}
		runner := SegmentRunner{
			MetricsCollector: &e.metricCollector,
			Logger:           &e.logger,
		}
		err = runner.Run(executionSegment, request, global)
		if err != nil {
			fmt.Errorf("Error running segment %s", err)
		}
		executionSegment = executionSegment.Next
	}
	return nil
}

func (e *Executor) executeRequest(request Request) Response {
	client := http.Client{}
	httpResponse := &http.Response{}
	startTime := time.Now()
	var err error
	switch request.Method {
	default:
		httpResponse, err = client.Get(request.URI)
	}
	if err != nil {
		return Response{Error: err}
	}
	response := Response{
		StatusCode: httpResponse.StatusCode,
		Duration:   time.Since(startTime),
	}
	responseData, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return Response{Error: err}
	}
	response.Body = string(responseData)
	return response
}

func validateResponse(response *http.Response, expectedResponse Response) (string, bool) {
	if response.StatusCode != expectedResponse.StatusCode {
		return fmt.Sprintf("http status %d, expected http status %d", response.StatusCode, expectedResponse.StatusCode), false
	}

	if len(expectedResponse.Headers) > 0 && len(response.Header) >= len(expectedResponse.Headers) {
		for _, header := range expectedResponse.Headers {
			if response.Header.Get(header.Get("key")) != header.Get("value") {
				return fmt.Sprintf("header %s, expected header %s", response.Header.Get(header.Get("key")), header.Get("value")), false
			}
		}
	}

	if len(expectedResponse.Cookies) > 0 && len(response.Cookies()) >= len(expectedResponse.Cookies) {
		for _, expectedCookie := range expectedResponse.Cookies {
			found := false
			for _, respCookie := range response.Cookies() {
				if respCookie.Name == expectedCookie.Name {
					found = true
					if respCookie.Value != expectedCookie.Value {
						return fmt.Sprintf("cookie %s=%s, expected cookie %s=%s", respCookie.Name, respCookie.Value, expectedCookie.Name, expectedCookie.Value), false
					}
				}
			}
			if !found {
				return fmt.Sprintf("cookie %s=%s not found in response", expectedCookie.Name, expectedCookie.Value), false
			}
		}
	}

	if expectedResponse.Body != "" {
		if response.Body == nil {
			return "body not matching, empty response", false
		}
		defer response.Body.Close()
		var expectedJSON, responseJSON interface{}

		err1 := json.Unmarshal([]byte(expectedResponse.Body), &expectedJSON)
		responseJSONString, err2 := io.ReadAll(response.Body)
		if err2 != nil {
			panic(err2)
		}
		err2 = json.Unmarshal(responseJSONString, &responseJSON)
		if err1 != nil || err2 != nil {
			return "Unable to parse string", false // Invalid JSON
		}

		if !reflect.DeepEqual(expectedJSON, responseJSON) {
			return "JSON response does not match expected response", false
		}
	}

	return "", true
}
