package goload

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
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
	Log   LogConfig
	Tests Test
}

func (e Executor) Execute() {
}

func Execute(config Config) {

	fmt.Print(logo + "\n\n\n")
	printConfig(config)
	fmt.Println("Starting execution...")
	logFilename := "execution-" + time.Now().Format("2006-01-02-15:04:05") + ".log"
	fmt.Println("creating log file : " + logFilename)
	if config.LogOutputPath != "" {
		err := os.MkdirAll(config.LogOutputPath, 0755)
		if err != nil {
			panic(err)
		}
	}
	file, err := os.Create(filepath.Join(config.LogOutputPath, logFilename))
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)
	if err != nil {
		panic(err)
	}
	safeWriter := SafeFileWriter{file: file, mu: sync.Mutex{}}
	safeWriter.file.WriteString(logo + "\n\n\n")
	_, err = safeWriter.file.WriteString("Starting execution at " + time.Now().Format("15:04:05") + "\n")
	if err != nil {
		return
	}

	var waitingGroup sync.WaitGroup
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	timepointExecutionStats := make([]TimepointExecutionStat, len(config.Timepoints), len(config.Timepoints))
	for timepointIndex := 0; timepointIndex < len(config.Timepoints); timepointIndex++ {
		fmt.Println("════════════════════════════════════════════════════")
		fmt.Print("Starting timepoint " + fmt.Sprint(timepointIndex) + " at " + time.Now().Format("15:04:05"))
		startTime := time.Now()
		timepointExecutionStats[timepointIndex] = TimepointExecutionStat{
			ExecutionTimepoint: config.Timepoints[timepointIndex],
		}
		for j := 0; j < config.Timepoints[timepointIndex].TargetVu; j++ {
			waitingGroup.Add(1)
			go execute(j, timepointIndex, config, &waitingGroup, httpClient, &safeWriter, &timepointExecutionStats[timepointIndex])
		}
		for {
			endTime := time.Now()
			if endTime.After(startTime.Add(config.Timepoints[timepointIndex].Duration)) || endTime.Equal(startTime.Add(config.Timepoints[timepointIndex].Duration)) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		fmt.Println("completed at " + time.Now().Format("15:04:05"))
	}

	waitingGroup.Wait()
	fmt.Println("════════════════════════════════════════════════════")
	fmt.Println("Execution completed at " + time.Now().Format("15:04:05"))
	fmt.Println("════════════════════════════════════════════════════")
	for i := 0; i < len(timepointExecutionStats); i++ {
		fmt.Println("\nSummary for timepoint " + fmt.Sprint(i) + " : \n")
		fmt.Println(timepointExecutionStats[i].printSuccessStats())
		fmt.Println(timepointExecutionStats[i].printResponseTimeStats())
		fmt.Println("════════════════════════════════════════════════════")
	}
}

func execute(vuId int, timepointId int, config Config, waitingGroup *sync.WaitGroup, httpClient *http.Client, safeFileWriter *SafeFileWriter, stat *TimepointExecutionStat) {
	defer waitingGroup.Done()
	endTime := time.Now().Add(config.Timepoints[timepointId].Duration)
	for {
		startTime := time.Now()
		if startTime.After(endTime) || startTime.Equal(endTime) {
			break
		}
		var resp *http.Response
		var respErr error
		if config.Request.Method == "" {
			panic("provide a valid http method")
		}
		req, reqErr := http.NewRequest(string(config.Request.Method), config.Request.URI, nil)
		if reqErr != nil {
			panic(reqErr)
		}
		req.Header.Set("User-Agent", string(config.Request.UserAgent))
		resp, respErr = httpClient.Do(req)
		if respErr != nil {
			panic(respErr)
		}
		defer resp.Body.Close()
		respTime := time.Since(startTime)
		stat.ResponseTime = append(stat.ResponseTime, float32(respTime.Seconds()))
		reason, match := validateResponse(resp, config.Response)
		stat.Success = append(stat.Success, match)
		var validString string
		if match == true {
			validString = "MATCH"
		} else {
			validString = "NO MATCH - " + reason
		}
		safeFileWriter.write(fmt.Sprintf("[%s][vu-%d] %s::%s resp(%.2fs) status(%d) %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			vuId,
			config.Request.Method,
			config.Request.URI,
			respTime.Seconds(),
			resp.StatusCode,
			validString))
	}
}

func printConfig(config Config) {
	fmt.Println("Executing loading test for the following parser :")
	fmt.Printf("Method: %v\n", config.Request.Method)
	fmt.Printf("URI: %s\n", config.Request.URI)
	fmt.Printf("Logging enabled: %v\n", config.Log)
	fmt.Println("Logging output path: " + config.LogOutputPath)
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
