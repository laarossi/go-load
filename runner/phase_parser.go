package runner

import "fmt"

func ResolvePhase(phase Phase) (*Segment, error) {
	if phase.SingleRequest != nil && *phase.SingleRequest {
		return &Segment{
			Duration:  nil,
			TargetVUs: 1,
		}, nil
	}

	if phase.Duration == nil && phase.SingleRequest == nil {
		return &Segment{}, fmt.Errorf("duration not specified")
	}

	if phase.Duration != nil && phase.SingleRequest != nil {
		return &Segment{}, fmt.Errorf("duration parameter cannot be used with single_request")
	}

	if phase.TargetVUs == nil {
		return &Segment{}, fmt.Errorf("you must specify target_vus")
	}

	if phase.Increment != nil && phase.TargetVUs == nil {
		return &Segment{}, fmt.Errorf("you must specify target_vus while using increment")
	}

	if (phase.IncrementVus != nil && phase.Increment == nil) || (phase.IncrementVus == nil && phase.Increment != nil) {
		return &Segment{}, fmt.Errorf("you must specify increment_vus while using increment")
	}

	return parsePhase(phase)

}

func parsePhase(phase Phase) (*Segment, error) {
	if phase.SingleRequest != nil && *phase.SingleRequest {
		return parseSingleRequestPhase(phase)
	}

	if phase.Increment != nil {
		return parseIncrementalPhase(phase)
	}

	return &Segment{
		Duration:  phase.Duration,
		TargetVUs: *phase.TargetVUs,
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
	incrementDuration := phase.Increment
	length := int(phase.Duration.Seconds() / phase.Increment.Seconds())
	var previousSegment *Segment
	for i := 0; i < length; i++ {
		segment := Segment{
			Duration:  incrementDuration,
			TargetVUs: *phase.TargetVUs + i**phase.IncrementVus,
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
