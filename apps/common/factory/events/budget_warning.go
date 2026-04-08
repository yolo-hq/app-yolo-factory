package events

import "github.com/yolo-hq/yolo/core/event"

// BudgetWarningEvent is emitted when a project approaches its budget limit.
type BudgetWarningEvent struct {
	event.CustomEvent[BudgetWarningPayload]
}

// BudgetWarningPayload carries budget warning data.
type BudgetWarningPayload struct {
	ProjectID  string  `entity:"Project"`
	Spent      float64 `json:"spent"`
	Limit      float64 `json:"limit"`
	Percentage float64 `json:"percentage"`
}
