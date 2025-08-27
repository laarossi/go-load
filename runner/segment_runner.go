package runner

import (
	"fmt"
	"goload/client"
	"goload/types"
	"goload/utils"
	"net/http"
	"sync"
	"time"
)

type Segment struct {
	TargetVUs int
	Duration  *time.Duration
	Request   *types.HTTPRequest
	Next      *Segment
}

type SegmentExecutionMetrics struct {
	RequestMetric types.RequestMetric
	NetworkMetric types.NetworkMetric
	CheckMetric   types.CheckMetric
}

type SegmentRunner struct {
	MetricsCollector *utils.MetricsCollector
	Logger           *utils.Logger
	Client           client.Client
}

func (runner *SegmentRunner) Run(segment *Segment, httpRequest types.HTTPRequest, global *Global) error {
	if segment.Request != nil {
		httpRequest = *segment.Request
	}

	var wg sync.WaitGroup
	startTIme := time.Now()
	first := true
	httpClient := &client.Client{
		HttpClient: &http.Client{},
	}
	for {
		if (segment.Duration == nil && !first) || (segment.Duration != nil && time.Since(startTIme) >= *segment.Duration) {
			break
		}
		first = false
		for i := 0; i < segment.TargetVUs; i++ {
			wg.Add(1)
			go func() {
				request, err := client.CreateRequest(httpRequest)
				if err != nil {
					_ = fmt.Errorf("error creating the httpRequest: %s\n", err)
				}
				response, err := httpClient.ExecuteRequest(request)
				runner.Logger.LogResponse(*response)
				err = runner.MetricsCollector.IngestRequestMetric(*response.RequestMetric)
				if err != nil {
					fmt.Printf("error ingesting request metric: %s\n", err)
				}
				if global.ThinkTime != nil {
					time.Sleep(*global.ThinkTime)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
	return nil
}
