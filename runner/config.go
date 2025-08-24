package runner

import (
	"goload/client"
	"time"
)

type Collection struct {
	Name  string `yaml:"name"`
	Tests []Test `yaml:"tests"`
}

type Test struct {
	Name    *string        `yaml:"name"`
	Global  *Global        `yaml:"global,omitempty"`
	Request client.Request `yaml:"request"`
	Phases  []Phase        `yaml:"phases"`
}
type Global struct {
	Timeout      time.Duration `yaml:"timeout,omitempty"` // e.g., "30s"
	Retries      int           `yaml:"retries,omitempty"` // Number of retries per request
	RetriesDelay time.Duration `yaml:"retries_delay,omitempty"`
	ThinkTime    time.Duration `yaml:"think_time,omitempty"` // Delay between requests per VU
}
type Phase struct {
	Name          *string         `yaml:"name"`
	SingleRequest *bool           `yaml:"single_request,omitempty"`
	Duration      *time.Duration  `yaml:"duration,omitempty"`
	Increment     *time.Duration  `yaml:"increment,omitempty"`
	IncrementVus  *int            `yaml:"increment_vus,omitempty"`
	TargetVUs     *int            `yaml:"target_vus,omitempty"`
	Request       *client.Request `yaml:"request"`
}

type LogConfig struct {
	LogOutputPath  string `yaml:"logOutputPath"`
	FilenameFormat string `yaml:"filenameFormat"`
}
