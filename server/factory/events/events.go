package events

import (
	"encoding/json"
	"log"
)

// Emit logs a structured event. Will be replaced with YOLO event system when wired.
func Emit(eventType string, payload any) {
	data, _ := json.Marshal(payload)
	log.Printf("[event] %s %s", eventType, string(data))
}

const (
	TaskStarted         = "factory.task.started"
	TaskCompleted       = "factory.task.completed"
	TaskFailed          = "factory.task.failed"
	PRDPlanningComplete = "factory.prd.planning_complete"
	PRDCompleted        = "factory.prd.completed"
	PRDFailed           = "factory.prd.failed"
	QuestionNeedsHuman  = "factory.question.needs_human"
	BudgetWarning       = "factory.budget.warning"
	BudgetExceeded      = "factory.budget.exceeded"
	SentinelBuildBroken = "factory.sentinel.build_broken"
	SentinelSecurityVuln = "factory.sentinel.security_vuln"
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

// SentinelPayload is sent for sentinel findings.
type SentinelPayload struct {
	ProjectName string `json:"project_name"`
	Error       string `json:"error"`
	Severity    string `json:"severity"`
}
