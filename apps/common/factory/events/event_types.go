package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// PRDCompletedEvent is emitted when all PRD tasks complete successfully.
type PRDCompletedEvent struct {
	event.EntityEvent[entities.PRD]
}

// PRDFailedEvent is emitted when a PRD fails.
type PRDFailedEvent struct {
	event.FailedEvent[entities.PRD]
}

// PRDPlanningCompleteEvent is emitted when PRD planning finishes.
type PRDPlanningCompleteEvent struct {
	event.EntityEvent[entities.PRD]
}

// QuestionNeedsHumanEvent is emitted when a question requires human input.
type QuestionNeedsHumanEvent struct {
	event.EntityEvent[entities.Question]
}

// TaskCompletedEvent is emitted when a task completes successfully.
type TaskCompletedEvent struct {
	event.EntityEvent[entities.Task]
}

// TaskFailedEvent is emitted when a task fails.
type TaskFailedEvent struct {
	event.FailedEvent[entities.Task]
}

// TaskStartedEvent is emitted when a task begins execution.
type TaskStartedEvent struct {
	event.EntityEvent[entities.Task]
}

// TaskBlockedEvent is emitted when a task cannot continue without human input.
type TaskBlockedEvent struct {
	event.FailedEvent[entities.Task]
}

// BudgetExceededEvent is emitted when a project exceeds its budget.
type BudgetExceededEvent struct {
	event.CustomEvent[BudgetExceededPayload]
}

// BudgetWarningEvent is emitted when a project approaches its budget limit.
type BudgetWarningEvent struct {
	event.CustomEvent[BudgetWarningPayload]
}

// SentinelBuildBrokenEvent is emitted when sentinel detects a broken build.
type SentinelBuildBrokenEvent struct {
	event.CustomEvent[SentinelPayload]
}

// SentinelSecurityVulnEvent is emitted when sentinel detects a security vulnerability.
type SentinelSecurityVulnEvent struct {
	event.CustomEvent[SentinelPayload]
}
