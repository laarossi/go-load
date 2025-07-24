package goload

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	logFilename := "execution-" + time.Now().Format("2006-08-12-15-04-30") + ".log"
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
	for i := 0; i < config.Vus; i++ {
		waitingGroup.Add(1)
		go execute(i, config, &waitingGroup, httpClient, &safeWriter, &respStats)
	}

	waitingGroup.Wait()

	sort.Slice(respStats, func(i, j int) bool { return respStats[i] < respStats[j] })
	percentiles := []float64{70, 80, 90, 95, 99}
	for _, p := range percentiles {
		idx := int(float64(len(respStats)-1) * p / 100)
		fmt.Printf("P%.0f: %.3fs\n", p, respStats[idx])
	}
}

func execute(vuId int, config Config, waitingGroup *sync.WaitGroup, httpClient *http.Client, safeFileWriter *SafeFileWriter, respStats *[]float32) {
	fmt.Println("Executing VU[" + fmt.Sprint(vuId) + "]")
	defer waitingGroup.Done()
	endTime := time.Now().Add(config.Execution.Duration)
	for {
		startTime := time.Now()
		if startTime.After(endTime) || startTime.Equal(endTime) {
			fmt.Println("VU[" + fmt.Sprint(vuId) + "] finished")
			break
		}
		var resp *http.Response
		var respErr error
		if config.Method == "" {
			panic("provide a valid http method")
		}
		req, reqErr := http.NewRequest(string(config.Method), config.URI, nil)
		if reqErr != nil {
			panic(reqErr)
		}
		req.Header.Set("User-Agent", config.UserAgent)
		resp, respErr = httpClient.Do(req)
		if respErr != nil {
			panic(respErr)
		}
		defer resp.Body.Close()
		respTime := time.Since(startTime)
		*respStats = append(*respStats, float32(respTime.Seconds()))
		safeFileWriter.write(fmt.Sprintf("[%s][vu-%d][%s] resp( %.3fs) status(%d)\n",
			time.Now().Format("2006-01-02 15:04:05"),
			vuId,
			config.Method,
			respTime,
			resp.StatusCode))
	}
}

func printConfig(config Config) {
	fmt.Println("Executing loading test for the following config :")
	fmt.Printf("Method: %v\n", config.Method)
	fmt.Printf("URI: %s\n", config.URI)
	fmt.Printf("VUs: %d\n", config.Vus)
	fmt.Printf("Logging enabled: %v\n", config.Log)
	fmt.Printf("Execution timepoint:\n")
	fmt.Printf("  Duration: %v\n", config.Execution.Duration)
	fmt.Printf("  Initial VUs: %d\n", config.Execution.InitialVUs)
}
