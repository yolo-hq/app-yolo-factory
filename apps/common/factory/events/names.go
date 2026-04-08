package events

// TaskPayload is sent with task failed events (carries transient data not on entity).
type TaskPayload struct {
	TaskID      string  `json:"task_id"`
	Title       string  `json:"title"`
	ProjectName string  `json:"project_name"`
	CostUSD     float64 `json:"cost_usd"`
	Error       string  `json:"error,omitempty"`
}
