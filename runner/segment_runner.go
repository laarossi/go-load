package runner

import (
	"goload/logger"
	"goload/metrics"
	"time"
)

type ExecutionSegment struct {
	TargetVUs int
	Duration  *time.Duration
	Next      *ExecutionSegment
}

type SegmentRunner struct {
	MetricsCollector *metrics.MetricsCollector
	Logger           *logger.Logger
}

func (executor *SegmentRunner) Run(segment ExecutionSegment) error {
	return nil
}
