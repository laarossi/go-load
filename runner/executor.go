package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Executor struct{}

func (e Executor) execute(config Config) {
	printConfig(config)
	logFilename := "execution-" + time.Now().Format("DD-MM-YYYY-HH-mm-ss") + ".log"
	fmt.Println("creating log file : " + logFilename)
	file, err := os.Create(config.LogOutputPath + "/" + logFilename)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)
	if err == nil {
		panic(err)
	}
	_, err = file.WriteString("Starting execution at " + time.Now().Format("15:04:05") + "\n")
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
	for i := 0; i < config.Vus; i++ {
		waitingGroup.Add(1)
		go execute(config, &waitingGroup, httpClient)
	}
}

func execute(config Config, waitingGroup *sync.WaitGroup, httpClient *http.Client) {
	defer waitingGroup.Done()
	req, err := http.NewRequest(string(config.Method), config.URI, nil)
	if err != nil {
		panic(err)
	}
	httpResponse, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer httpResponse.Body.Close()
	_, err = httpResponse.Body.Read(make([]byte, 1024))
	if err != nil {
		panic(err)
	}

}

func printConfig(config Config) {
	fmt.Println("Executing loading test for the following config :")
	fmt.Printf("Method: %v\n", config.Method)
	fmt.Printf("URI: %s\n", config.URI)
	fmt.Printf("VUs: %d\n", config.Vus)
	fmt.Printf("Logging enabled: %v\n", config.Log)
	fmt.Printf("Execution timepoint:\n")
	fmt.Printf("  Duration: %v\n", config.Execution.duration)
	fmt.Printf("  Initial VUs: %d\n", config.Execution.initialVus)
}
