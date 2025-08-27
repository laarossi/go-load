package runner

import (
	"goload/types"
	"strconv"
	"strings"
	"time"
)

type Collection struct {
	Name  string `yaml:"name"`
	Tests []Test `yaml:"tests"`
}

type Test struct {
	Name    string            `yaml:"name"`
	Global  *Global           `yaml:"global,omitempty"`
	Request types.HTTPRequest `yaml:"request"`
	Phases  []Phase           `yaml:"phases"`
}
type Global struct {
	Timeout      time.Duration  `yaml:"timeout,omitempty"` // e.g., "30s"
	Retries      int            `yaml:"retries,omitempty"` // Number of retries per request
	RetriesDelay time.Duration  `yaml:"retries_delay,omitempty"`
	ThinkTime    *time.Duration `yaml:"think_time,omitempty"` // Delay between requests per VU
}
type Phase struct {
	Name          string             `yaml:"name"`
	SingleRequest bool               `yaml:"single_request,omitempty"`
	Duration      string             `yaml:"duration,omitempty"`
	Increment     string             `yaml:"increment,omitempty"`
	IncrementVus  int                `yaml:"increment_vus,omitempty"`
	TargetVUs     int                `yaml:"target_vus,omitempty"`
	Request       *types.HTTPRequest `yaml:"request"`
}

func (p Phase) String() string {
	var result []string

	if p.Name != "" {
		result = append(result, "name:"+p.Name)
	}
	if p.SingleRequest == true {
		result = append(result, "single_request:"+strconv.FormatBool(p.SingleRequest))
	}
	if p.Duration != "" {
		result = append(result, "duration:"+p.Duration)
	}
	if p.Increment != "" {
		result = append(result, "increment:"+p.Increment)
	}
	if p.IncrementVus != 0 {
		result = append(result, "increment_vus:"+strconv.Itoa(p.IncrementVus))
	}
	if p.TargetVUs != 0 {
		result = append(result, "target_vus:"+strconv.Itoa(p.TargetVUs))
	}

	return strings.Join(result, " | ")
}
