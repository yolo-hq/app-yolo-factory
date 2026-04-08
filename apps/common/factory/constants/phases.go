package constants

// StepPhase identifies the phase of a workflow step.
type StepPhase string

const (
	PhasePlan      StepPhase = "plan"
	PhaseImplement StepPhase = "implement"
	PhaseTest      StepPhase = "test"
	PhaseLint      StepPhase = "lint"
	PhaseAudit     StepPhase = "audit"
	PhaseReview    StepPhase = "review"
)
