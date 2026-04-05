package entities

import (
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Review struct {
	bun.BaseModel `bun:"table:factory_reviews"`
	entity.BaseEntity
	RunID           string  `json:"runId" bun:"run_id,notnull" fake:"rel:Run"`
	TaskID          string  `json:"taskId" bun:"task_id,notnull" fake:"rel:Task"`
	SessionID       string  `json:"sessionId" bun:"session_id" fake:"-"`
	Model           string  `json:"model" bun:"model,notnull" fake:"oneof:sonnet,opus,haiku"`
	Verdict         string  `json:"verdict" bun:"verdict,notnull" fake:"oneof:pass,fail,retry"`
	Reasons         string  `json:"reasons" bun:"reasons,default:'[]'" fake:"-"`
	AntiPatterns    string  `json:"antiPatterns" bun:"anti_patterns,default:'[]'" fake:"-"`
	CriteriaResults string  `json:"criteriaResults" bun:"criteria_results,notnull" fake:"sentence:15"`
	Suggestions     string  `json:"suggestions" bun:"suggestions,default:'[]'" fake:"-"`
	CostUSD         float64 `json:"costUsd" bun:"cost_usd,default:0" fake:"float:0,5"`

	// Relations
	Run  *Run  `json:"run,omitempty" bun:"-" yolo:"rel:belongs_to,fk:run_id"`
	Task *Task `json:"task,omitempty" bun:"-" yolo:"rel:belongs_to,fk:task_id"`
}

func (Review) TableName() string  { return "factory_reviews" }
func (Review) EntityName() string { return "Review" }
