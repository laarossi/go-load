package runner

import (
	fmt "fmt"
	"goload/types"
	"goload/utils"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Executor struct {
	collection      Collection
	logger          utils.Logger
	metricCollector utils.MetricsCollector
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
	newLogger, _ := utils.NewLogger("logs")
	e.logger = *newLogger
	e.metricCollector = utils.MetricsCollector{
		Logger: *newLogger,
	}
	return nil
}

func (e *Executor) Execute() {
	err := e.metricCollector.Init()
	e.metricCollector.StartWorkers()
	if err != nil {
		_ = fmt.Errorf("error initializing metrics collector: %s", err)
	}
	_ = e.logger.Log(fmt.Sprintf("Executing %d tests", len(e.collection.Tests)))
	for _, test := range e.collection.Tests {
		if test.Name != nil {
			fmt.Println("parsing test configuration for ", *test.Name)
		} else {
			fmt.Println("parsing test configuration")
		}
		for i, phase := range test.Phases {
			if phase.Request == nil {
				phase.Request = &test.Request
			}
			_ = e.logger.LogSeparator()
			_ = e.logger.Log(fmt.Sprintf("Executing phase number : %d", i+1))
			_ = e.logger.Log(phase.String())
			_ = e.logger.Log(phase.Request.Summary())
			_ = e.logger.LogSeparator()
			err := e.executePhase(phase, test.Request, test.Global)
			if err != nil {
				_ = fmt.Errorf("failed to execute phase: %s", err)
			}
		}
	}
	e.metricCollector.StopWorkers()
	e.metricCollector.LogRequestsStats()
}

func (e *Executor) executePhase(phase Phase, request types.HTTPRequest, global *Global) error {
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
