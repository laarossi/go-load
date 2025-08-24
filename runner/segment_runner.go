package runner

import (
	"goload/client"
	"goload/logger"
	"goload/metrics"
	"net/http"
	"sync"
	"time"
)

type Segment struct {
	TargetVUs int
	Duration  *time.Duration
	Request   *client.Request
	Next      *Segment
}

type SegmentExecutionMetrics struct {
	RequestMetric metrics.RequestMetric
	NetworkMetric metrics.NetworkMetric
	CheckMetric   metrics.CheckMetric
}

type SegmentRunner struct {
	MetricsCollector *metrics.MetricsCollector
	Logger           *logger.Logger
	Client           client.Client
}

func (runner *SegmentRunner) Run(segment *Segment, request client.Request, global *Global) error {
	if segment.Request != nil {
		request = *segment.Request
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
			go func() {
				requestMetrics, networkMetrics, err := httpClient.DoRequest(string(request.Method), request.URI, &client.RequestOptions{
					Headers: request.Headers,
					Cookies: request.Cookies,
					Body:    &request.Body,
				})
				if err != nil {
					return
				}
				err = runner.MetricsCollector.IngestRequestMetric(*requestMetrics)
				if err != nil {
					return
				}
				err = runner.MetricsCollector.IngestNetworkMetric(*networkMetrics)
				if err != nil {
					return
				}
			}()
			wg.Add(1)
		}
		wg.Wait()
	}
	return nil
}
