package goload

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"time"
)

type SafeFileWriter struct {
	file *os.File
	mu   sync.Mutex
}

func (sw *SafeFileWriter) write(data string) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.file.WriteString(data)
}

func Execute(config Config) {
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

	respStats := []float32{}
	success := []bool{}
	for timepointIndex := 0; timepointIndex < len(config.Timepoints); timepointIndex++ {
		fmt.Println("Starting timepoint " + fmt.Sprint(timepointIndex) + " at " + time.Now().Format("15:04:05"))
		startTime := time.Now()
		for j := 0; j < config.Timepoints[timepointIndex].TargetVu; j++ {
			waitingGroup.Add(1)
			go execute(j, timepointIndex, config, &waitingGroup, httpClient, &safeWriter, &respStats, &success)
		}
		for {
			endTime := time.Now()
			if endTime.After(startTime.Add(config.Timepoints[timepointIndex].Duration)) || endTime.Equal(startTime.Add(config.Timepoints[timepointIndex].Duration)) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	waitingGroup.Wait()
	sort.Slice(respStats, func(i, j int) bool { return respStats[i] < respStats[j] })
	percentiles := []float64{70, 80, 90, 95, 99}
	fmt.Println("Execution finished at " + time.Now().Format("15:04:05"))
	fmt.Println("Execution stats :")
	fmt.Println("Response time stats :")
	fmt.Println("   70th percentile : " + fmt.Sprint(percentileResponseTime(respStats, percentiles[0])))
	fmt.Println("   80th percentile : " + fmt.Sprint(percentileResponseTime(respStats, percentiles[1])))
	fmt.Println("   90th percentile : " + fmt.Sprint(percentileResponseTime(respStats, percentiles[2])))
	fmt.Println("   95th percentile : " + fmt.Sprint(percentileResponseTime(respStats, percentiles[3])))
	fmt.Println("   99th percentile : " + fmt.Sprint(percentileResponseTime(respStats, percentiles[4])))
	fmt.Println("Response success stats :")
	fmt.Println("   70th percentile : " + fmt.Sprint(percentileSuccess(success, percentiles[0])))
	fmt.Println("   80th percentile : " + fmt.Sprint(percentileSuccess(success, percentiles[1])))
	fmt.Println("   90th percentile : " + fmt.Sprint(percentileSuccess(success, percentiles[2])))
	fmt.Println("   95th percentile : " + fmt.Sprint(percentileSuccess(success, percentiles[3])))
	fmt.Println("   99th percentile : " + fmt.Sprint(percentileSuccess(success, percentiles[4])))
}

func percentileResponseTime(data []float32, percentile float64) float32 {
	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
	idx := int(float64(len(data)-1) * percentile / 100)
	return data[idx]
}

func percentileSuccess(data []bool, percentile float64) float64 {
	sort.Slice(data, func(i, j int) bool { return data[i] && !data[j] })
	idx := int(float64(len(data)-1) * percentile / 100)
	return float64(idx)
}

func execute(vuId int, timepointId int, config Config, waitingGroup *sync.WaitGroup, httpClient *http.Client, safeFileWriter *SafeFileWriter, respStats *[]float32, success *[]bool) {
	fmt.Println("Executing VU[" + fmt.Sprint(vuId) + "]")
	defer waitingGroup.Done()
	endTime := time.Now().Add(config.Timepoints[timepointId].Duration)
	for {
		startTime := time.Now()
		if startTime.After(endTime) || startTime.Equal(endTime) {
			fmt.Println("VU[" + fmt.Sprint(vuId) + "] finished executing timepoint " + fmt.Sprint(timepointId) + " at " + time.Now().Format("15:04:05"))
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
		*respStats = append(*respStats, float32(respTime.Seconds()))
		reason, match := validateResponse(resp, config.Response)
		*success = append(*success, match)
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
	fmt.Println("Executing loading test for the following config :")
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
