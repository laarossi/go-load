package runner

import (
	fmt "fmt"
	"goload/client"
	"goload/logger"
	"goload/metrics"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Executor struct {
	collection      Collection
	logger          logger.Logger
	metricCollector metrics.MetricsCollector
}

func LoadFromYaml(yamlFilePath string) (*Executor, error) {
	configPath := filepath.Join(yamlFilePath)
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return &Executor{}, err
	}

	var collection Collection
	err = yaml.Unmarshal(yamlFile, &collection)
	if err != nil {
		fmt.Printf("Error parsing YAML: %s\n", err)
		return &Executor{}, err
	}

	executor := Executor{
		collection: collection,
	}
	err = executor.load()
	if err != nil {
		return &Executor{}, err
	}
	return &executor, nil
}

func (e *Executor) load() error {
	newLogger, _ := logger.NewLogger("logs")
	e.logger = *newLogger
	e.metricCollector = metrics.MetricsCollector{
		LogDir: "logs",
	}
	return nil
}

func (e *Executor) Execute() {
	err := e.metricCollector.Init()
	e.metricCollector.StartWorkers()
	if err != nil {
		_ = fmt.Errorf("error initializing metrics collector: %s", err)
	}
	for _, test := range e.collection.Tests {
		if test.Name != nil {
			fmt.Println("parsing test configuration for ", *test.Name)
		} else {
			fmt.Println("parsing test configuration")
		}
		for _, phase := range test.Phases {
			err := e.executePhase(phase, test.Request, test.Global)
			if err != nil {
				_ = fmt.Errorf("failed to execute phase: %s", err)
			}
		}
	}
	e.metricCollector.StopWorkers()
	fmt.Println("\nResponse Time Stats:")
	percentiles := []float32{50, 75, 90, 95, 99}
	for _, p := range percentiles {
		e.metricCollector.PrintRequestLatencyPercentiles(p)
	}
}

func (e *Executor) executePhase(phase Phase, request client.HTTPRequest, global *Global) error {
	executionSegment, err := ResolvePhase(phase)
	if err != nil {
		fmt.Printf("Error resolving phase: %s\n", err)
		return nil
	}
	for {
		if executionSegment == nil {
			break
		}
		runner := SegmentRunner{
			MetricsCollector: &e.metricCollector,
			Logger:           &e.logger,
		}
		err = runner.Run(executionSegment, request, global)
		if err != nil {
			_ = fmt.Errorf("error running segment %s", err)
			continue
		}
		executionSegment = executionSegment.Next
	}
	return nil
}
