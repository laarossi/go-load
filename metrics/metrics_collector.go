package metrics

import (
	"fmt"
	"goload/utils"
	"io"
	"os"

	"sync"
	"time"
)

type MetricsCollector struct {
	InMemory         bool
	logWriter        io.Writer
	LogDir           string
	metricsMutex     []sync.Mutex
	RequestsMetrics  []RequestMetric
	NetworkMetrics   []NetworkMetric
	CheckMetrics     []CheckMetric
	MetricWorkerPool *utils.WorkerPool[MetricWorkerTask]
}

type MetricWorkerTask struct {
	TaskType string
	TaskData interface{}
}

func (collector *MetricsCollector) Init() error {
	if !collector.InMemory {
		collector.RequestsMetrics = make([]RequestMetric, 0)
		collector.NetworkMetrics = make([]NetworkMetric, 0)
		collector.CheckMetrics = make([]CheckMetric, 0)
		collector.metricsMutex = make([]sync.Mutex, 1)
		collector.metricsMutex[0] = sync.Mutex{}
		fileName := collector.LogDir + "/metrics-" + time.Now().Format("2006-01-02 15:04:05")
		if err := os.MkdirAll(collector.LogDir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %s", err)
		}
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("error creating metrics file: %s", err)
		}
		collector.logWriter = io.Writer(file)
	}
	collector.MetricWorkerPool = utils.NewWorkerPool[MetricWorkerTask](10, func(task MetricWorkerTask) {
		err := collector.MetricWorkerHandler(task)
		if err != nil {
			return
		}
	})
	return nil
}

func (collector *MetricsCollector) MetricWorkerHandler(task MetricWorkerTask) error {
	if task.TaskType == "request" {
		fmt.Println("request metric received")
	} else if task.TaskType == "network" {
		//
	} else if task.TaskType == "check" {
		// Define task
	} else {
		return fmt.Errorf("unknown task type: %s", task.TaskType)
	}
	return nil
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

func (collector *MetricsCollector) Close() {
	collector.MetricWorkerPool.Stop()
}
