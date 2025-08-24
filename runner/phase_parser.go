package runner

import "fmt"

func ResolvePhase(phase Phase) (ExecutionSegment, error) {
	if phase.SingleRequest != nil && *phase.SingleRequest {
		return ExecutionSegment{
			Duration:  nil,
			TargetVUs: 1,
		}, nil
	}

	if phase.Duration == nil && phase.SingleRequest == nil {
		return ExecutionSegment{}, fmt.Errorf("duration not specified")
	}

	if phase.Duration != nil && phase.SingleRequest != nil {
		return ExecutionSegment{}, fmt.Errorf("duration parameter cannot be used with single_request")
	}

	if phase.TargetVUs == nil {
		return ExecutionSegment{}, fmt.Errorf("you must specify target_vus")
	}

	if phase.Increment != nil && phase.TargetVUs == nil {
		return ExecutionSegment{}, fmt.Errorf("you must specify target_vus while using increment")
	}

	if (phase.IncrementVus != nil && phase.Increment == nil) || (phase.IncrementVus == nil && phase.Increment != nil) {
		return ExecutionSegment{}, fmt.Errorf("you must specify increment_vus while using increment")
	}

	return parsePhase(phase)

}

func parsePhase(phase Phase) (ExecutionSegment, error) {
	if phase.SingleRequest != nil && *phase.SingleRequest {
		return parseSingleRequestPhase()
	}

	if phase.Increment != nil {
		return parseIncrementalPhase(phase)
	}

	return ExecutionSegment{
		Duration:  phase.Duration,
		TargetVUs: *phase.TargetVUs,
		Next:      nil,
	}, nil
}

func parseSingleRequestPhase() (ExecutionSegment, error) {
	return ExecutionSegment{
		Duration:  nil,
		TargetVUs: 1,
	}, nil
}

func parseIncrementalPhase(phase Phase) (ExecutionSegment, error) {
	var headSegment ExecutionSegment
	incrementDuration := phase.Increment
	length := int(phase.Duration.Seconds() / phase.Increment.Seconds())
	var previousSegment *ExecutionSegment
	for i := 0; i < length; i++ {
		segment := ExecutionSegment{
			Duration:  incrementDuration,
			TargetVUs: *phase.TargetVUs + i**phase.IncrementVus,
		}

		if i == 0 {
			headSegment = segment
		}
		previousSegment.Next = &segment
		previousSegment = &segment

	}
	return headSegment, nil
}
