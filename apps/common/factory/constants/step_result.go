package constants

// Step result statuses — returned by executeStep and stored on Step.ResultStatus.
//
// done              → continue to next step; all clear.
// done_with_concerns → continue but log concerns; include in final review.
// needs_context     → trigger question escalation (existing flow).
// blocked           → stop run, create Question with "blocked" context.
// failed            → terminal failure, same as previous Failed=true behaviour.
const (
	StepResultDone             = "done"
	StepResultDoneWithConcerns = "done_with_concerns"
	StepResultNeedsContext     = "needs_context"
	StepResultBlocked          = "blocked"
	StepResultFailed           = "failed"
)

// RunStatusBlocked is the run status set when a task is waiting on human input.
// The value must stay in sync with the "blocked" enum value on the Run entity.
const RunStatusBlocked = "blocked"
