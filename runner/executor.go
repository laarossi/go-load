package goload

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

const logo = `
	 ██████╗  ██████╗ ██╗	  ██████╗  █████╗ ██████╗ 
	██╔════╝ ██╔═══██╗██║	 ██╔═══██╗██╔══██╗██╔══██╗
	██║  ███╗██║   ██║██║	 ██║   ██║███████║██║  ██║
	██║   ██║██║   ██║██║	 ██║   ██║██╔══██║██║  ██║
	╚██████╔╝╚██████╔╝███████╗╚██████╔╝██║  ██║██████╔╝
	 ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝ 
	════════════════════════════════════════════════════`

type Executor struct {
	Collection Collection
}

func (e *Executor) Execute() {
	fmt.Println(logo)
	for _, test := range e.Collection.Tests {
		e.executeTest(test)
	}
}

func (e *Executor) executeTest(test Test) {
	fmt.Println("configuring test: ", test.Name)
	for _, phase := range test.Phases {
		e.executePhase(phase, test.Request)
	}
}

func (e *Executor) executePhase(phase Phase, request Request) {
	fmt.Println("parsing phase configuration for ", phase.Name)
	if phase.SingleRequest != nil && *phase.SingleRequest {
		fmt.Println("executing single request")
		response := e.executeRequest(request)
		if response.Error != nil {
			fmt.Println("error executing request: ", response.Error)
		}
	} else {
		fmt.Println("executing multiple requests")
	}
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
