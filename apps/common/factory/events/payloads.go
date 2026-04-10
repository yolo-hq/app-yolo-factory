package events

// TaskPayload is sent with task failed events (carries transient data not on entity).
type TaskPayload struct {
	TaskID      string  `json:"task_id"`
	Title       string  `json:"title"`
	ProjectName string  `json:"project_name"`
	CostUSD     float64 `json:"cost_usd"`
	Error       string  `json:"error,omitempty"`
}

// BudgetExceededPayload carries budget exceeded data.
type BudgetExceededPayload struct {
	ProjectID  string  `entity:"Project"`
	Spent      float64 `json:"spent"`
	Limit      float64 `json:"limit"`
	Percentage float64 `json:"percentage"`
}

// BudgetWarningPayload carries budget warning data.
type BudgetWarningPayload struct {
	ProjectID  string  `entity:"Project"`
	Spent      float64 `json:"spent"`
	Limit      float64 `json:"limit"`
	Percentage float64 `json:"percentage"`
}

// SentinelPayload carries sentinel finding data.
type SentinelPayload struct {
	ProjectID string `entity:"Project"`
	Error     string `json:"error"`
	Severity  string `json:"severity"`
}
