package metrics

import (
	"fmt"
	"github.com/HdrHistogram/hdrhistogram-go"
	"goload/utils"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"time"
)

type MetricsCollector struct {
	logWriter                    io.Writer
	LogDir                       string
	requestLatencyHistogramMutex *sync.Mutex
	requestLatencyHistogram      *hdrhistogram.Histogram
	totalChacks                  int64
	totalRequests                int64
	totalFails                   int64
	totalSuccesses               int64
	MetricWorkerPool             *utils.WorkerPool[MetricWorkerTask]
}

type MetricWorkerTask struct {
	TaskType string
	TaskData interface{}
}

func (collector *MetricsCollector) Init() error {
	collector.requestLatencyHistogram = hdrhistogram.New(1, 60_000_000, 3)
	collector.requestLatencyHistogramMutex = &sync.Mutex{}
	fileName := collector.LogDir + "/metrics-" + time.Now().Format("2006-01-02 15:04:05")
	if err := os.MkdirAll(collector.LogDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %s", err)
	}
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating metrics file: %s", err)
	}
	collector.logWriter = io.Writer(file)

	collector.MetricWorkerPool = utils.NewWorkerPool[MetricWorkerTask](10, func(task MetricWorkerTask) {
		err := collector.metricWorkerHandler(task)
		if err != nil {
			return
		}
	})
	return nil
}

func (collector *MetricsCollector) metricWorkerHandler(task MetricWorkerTask) error {
	if task.TaskType == "request" {
		requestMetric := task.TaskData.(RequestMetric)
		atomic.AddInt64(&collector.totalRequests, 1)
		collector.requestLatencyHistogramMutex.Lock()
		err := collector.requestLatencyHistogram.RecordValue(requestMetric.Duration.Milliseconds())
		collector.requestLatencyHistogramMutex.Unlock()
		if err != nil {
			_ = fmt.Errorf("error recording request latency: %s", err)
		}
	} else if task.TaskType == "network" {
		//
	} else if task.TaskType == "check" {
		// Define task
	} else {
		return fmt.Errorf("unknown task type: %s", task.TaskType)
	}
	return nil
}

func (collector *MetricsCollector) PrintRequestLatencyPercentiles(percentile float32) {
	fmt.Printf("Request Latency Percentiles at %03.1f %% : %f\n", percentile, float64(collector.requestLatencyHistogram.ValueAtQuantile(float64(percentile))))
}

func (collector *MetricsCollector) IngestRequestMetric(metric RequestMetric) error {
	metricTask := MetricWorkerTask{
		TaskType: "request",
		TaskData: metric,
	}
	collector.MetricWorkerPool.AddTask(metricTask)
	return nil
}

func (collector *MetricsCollector) IngestNetworkMetric(metric NetworkMetric) error {
	metricTask := MetricWorkerTask{
		TaskType: "network",
		TaskData: metric,
	}
	collector.MetricWorkerPool.AddTask(metricTask)
	return nil
}

func (collector *MetricsCollector) IngestCheckMetric(metric CheckMetric) error {
	return nil
}

func (collector *MetricsCollector) StartWorkers() {
	collector.MetricWorkerPool.Start()
}

func (collector *MetricsCollector) StopWorkers() {
	collector.MetricWorkerPool.Stop()
}
