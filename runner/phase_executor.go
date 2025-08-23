package runner

const (
	ProfileSmoke  Profile = "smoke"
	ProfileSpike  Profile = "spike"
	ProfileLoad   Profile = "load"
	ProfileCustom Profile = "custom"
)

type PhaseExecutor interface {
	RunPhase() error
}

type SmokePhaseExecutor struct{}

func (e *SmokePhaseExecutor) RunPhase() error {
	return nil
}

type SpikePhaseExecutor struct{}

func (e *SpikePhaseExecutor) RunPhase() error {
	return nil
}

type LoadPhaseExecutor struct{}

func (e *LoadPhaseExecutor) RunPhase() error {
	return nil
}

type CustomPhaseExecutor struct{}

func (e *CustomPhaseExecutor) RunPhase() error {
	return nil
}
