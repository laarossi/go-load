package metrics

import (
	"fmt"
	"github.com/HdrHistogram/hdrhistogram-go"
	"goload/internal/logging"
	"goload/internal/worker"
	"goload/types"
	"sync"
)

type MetricsCollector struct {
	Logger                       logging.Logger
	requestLatencyHistogramMutex *sync.Mutex
	requestLatencyHistogram      *hdrhistogram.Histogram
	totalChacks                  int64
	totalRequests                int64
	totalFails                   int64
	totalSuccesses               int64
	MetricWorkerPool             *worker.WorkerPool[MetricWorkerTask]
}

type MetricWorkerTask struct {
	TaskType string
	TaskData interface{}
}

func (collector *MetricsCollector) Init() error {
	collector.requestLatencyHistogram = hdrhistogram.New(1, 60_000_000, 3)
	collector.requestLatencyHistogramMutex = &sync.Mutex{}
	collector.MetricWorkerPool = worker.NewWorkerPool[MetricWorkerTask](10, func(task MetricWorkerTask) {
		err := collector.metricWorkerHandler(task)
		if err != nil {
			return
		}
	})
	return nil
}

func (collector *MetricsCollector) metricWorkerHandler(task MetricWorkerTask) error {
	if task.TaskType == "request" {
		requestMetric := task.TaskData.(types.RequestMetric)
		collector.requestLatencyHistogramMutex.Lock()
		err := collector.requestLatencyHistogram.RecordValue(requestMetric.Duration.Milliseconds())
		collector.totalRequests++
		if requestMetric.StatusCode >= 200 && requestMetric.StatusCode < 300 {
			collector.totalSuccesses++
		} else {
			collector.totalFails++
		}
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

func (collector *MetricsCollector) IngestRequestMetric(metric types.RequestMetric) error {
	metricTask := MetricWorkerTask{
		TaskType: "request",
		TaskData: metric,
	}
	collector.MetricWorkerPool.AddTask(metricTask)
	return nil
}

func (collector *MetricsCollector) IngestNetworkMetric(metric types.NetworkMetric) error {
	metricTask := MetricWorkerTask{
		TaskType: "network",
		TaskData: metric,
	}
	collector.MetricWorkerPool.AddTask(metricTask)
	return nil
}

func (collector *MetricsCollector) IngestCheckMetric(metric types.CheckMetric) error {
	return nil
}

func (collector *MetricsCollector) LogRequestsStats() {
	collector.requestLatencyHistogramMutex.Lock()
	defer collector.requestLatencyHistogramMutex.Unlock()

	table := fmt.Sprintf("+-----------------+-----------+\n")
	table += fmt.Sprintf("| Metric		  | Value	 |\n")
	table += fmt.Sprintf("+-----------------+-----------+\n")
	table += fmt.Sprintf("| Total Requests  | %-9d |\n", collector.totalRequests)
	table += fmt.Sprintf("| Total Successes | %-9d |\n", collector.totalSuccesses)
	table += fmt.Sprintf("| Total Fails	 | %-9d |\n", collector.totalFails)
	table += fmt.Sprintf("+-----------------+-----------+\n")
	table += fmt.Sprintf("\nLatency Percentiles (ms):\n")
	table += fmt.Sprintf("+------------+-----------+\n")
	table += fmt.Sprintf("| Percentile | Latency   |\n")
	table += fmt.Sprintf("+------------+-----------+\n")

	percentiles := []float64{60, 70, 75, 80, 85, 90, 95, 97, 98, 99}
	for _, p := range percentiles {
		table += fmt.Sprintf("| p%-9.1f | %-9.1f |\n", p, float64(collector.requestLatencyHistogram.ValueAtQuantile(p)))
	}
	table += fmt.Sprintf("+------------+-----------+\n")

	collector.Logger.LogWithoutDate(table)
}

func (collector *MetricsCollector) StartWorkers() {
	collector.MetricWorkerPool.Start()
}

func (collector *MetricsCollector) StopWorkers() {
	collector.MetricWorkerPool.Stop()
}
