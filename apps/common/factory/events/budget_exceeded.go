package events

import "github.com/yolo-hq/yolo/core/event"

// BudgetExceededEvent is emitted when a project exceeds its budget.
type BudgetExceededEvent struct {
	event.CustomEvent[BudgetExceededPayload]
}

// BudgetExceededPayload carries budget exceeded data.
type BudgetExceededPayload struct {
	ProjectID  string  `entity:"Project"`
	Spent      float64 `json:"spent"`
	Limit      float64 `json:"limit"`
	Percentage float64 `json:"percentage"`
}
