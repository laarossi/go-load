package runner

import (
	"fmt"
	"time"
)

func ResolvePhase(phase Phase) (*Segment, error) {
	if phase.SingleRequest {
		return &Segment{
			Duration:  nil,
			TargetVUs: 1,
		}, nil
	}

	if phase.Duration == "" && !phase.SingleRequest {
		return &Segment{}, fmt.Errorf("duration not specified")
	}

	if phase.Duration != "" && phase.SingleRequest {
		return &Segment{}, fmt.Errorf("duration parameter cannot be used with single_request")
	}

	if phase.TargetVUs == 0 {
		return &Segment{}, fmt.Errorf("you must specify target_vus")
	}

	if phase.Increment != "" && phase.TargetVUs == 0 {
		return &Segment{}, fmt.Errorf("you must specify target_vus while using increment")
	}

	if (phase.IncrementVus != 0 && phase.Increment == "") || (phase.IncrementVus == 0 && phase.Increment != "") {
		return &Segment{}, fmt.Errorf("you must specify increment_vus while using increment")
	}

	return parsePhase(phase)

}

func parsePhase(phase Phase) (*Segment, error) {
	if phase.SingleRequest != false && phase.SingleRequest {
		return parseSingleRequestPhase(phase)
	}

	if phase.Increment != "" {
		return parseIncrementalPhase(phase)
	}

	phaseDuration, err := time.ParseDuration(phase.Duration)
	if err != nil {
		return nil, err
	}

	return &Segment{
		Duration:  &phaseDuration,
		TargetVUs: phase.TargetVUs,
		Request:   phase.Request,
		Next:      nil,
	}, nil
}

func parseSingleRequestPhase(phase Phase) (*Segment, error) {
	return &Segment{
		Duration:  nil,
		TargetVUs: 1,
		Request:   phase.Request,
	}, nil
}

func parseIncrementalPhase(phase Phase) (*Segment, error) {
	var headSegment Segment
	phaseDuration, err := time.ParseDuration(phase.Duration)
	if err != nil {
		return nil, err
	}
	incrementDuration, err := time.ParseDuration(phase.Duration)
	if err != nil {
		return nil, err
	}
	length := int(phaseDuration.Seconds() / incrementDuration.Seconds())
	var previousSegment *Segment
	for i := 0; i < length; i++ {
		segment := Segment{
			Duration:  &incrementDuration,
			TargetVUs: phase.TargetVUs + i*phase.IncrementVus,
			Request:   phase.Request,
		}

		if i == 0 {
			headSegment = segment
			previousSegment = &headSegment
			continue
		}
		previousSegment.Next = &segment
		previousSegment = &segment

	}
	return &headSegment, nil
}
