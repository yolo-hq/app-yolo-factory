package events

// FailedEvent name constants — not auto-generated (FailedEvent type not yet supported by generator).
const (
	TaskFailedName = "task.failed"
	PRDFailedName  = "prd.failed"
)

// TaskPayload is sent with task lifecycle events.
type TaskPayload struct {
	TaskID      string  `json:"task_id"`
	Title       string  `json:"title"`
	ProjectName string  `json:"project_name"`
	CostUSD     float64 `json:"cost_usd"`
	Error       string  `json:"error,omitempty"`
}

// PRDPayload is sent with PRD lifecycle events.
type PRDPayload struct {
	PRDID        string  `json:"prd_id"`
	Title        string  `json:"title"`
	TaskCount    int     `json:"task_count"`
	TotalCostUSD float64 `json:"total_cost_usd"`
}

// QuestionPayload is sent when a question needs human input.
type QuestionPayload struct {
	QuestionID string `json:"question_id"`
	TaskID     string `json:"task_id"`
	Body       string `json:"body"`
	Context    string `json:"context"`
}

// BudgetPayload is sent for budget warnings and exceeded events.
type BudgetPayload struct {
	ProjectName string  `json:"project_name"`
	Spent       float64 `json:"spent"`
	Limit       float64 `json:"limit"`
	Percentage  float64 `json:"percentage"`
}
