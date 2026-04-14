package entities

import (
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Suggestion struct {
	bun.BaseModel `bun:"table:factory_suggestions"`
	entity.BaseEntity
	ProjectID       string `json:"project_id" bun:"project_id,notnull" fake:"rel:Project"`
	Source          string `json:"source" bun:"source,notnull" fake:"oneof:review,sentinel,advisor,manual"`
	Category        string `json:"category" bun:"category,notnull" fake:"oneof:bug,feature,refactor,test,docs" enum:"optimization,refactoring,tech_debt,security,new_feature,pattern_extraction,bug_fix"`
	Title           string `json:"title" bun:"title,notnull" fake:"sentence:6"`
	Body            string `json:"body" bun:"body,notnull" fake:"sentence:20"`
	Priority        string `json:"priority" bun:"priority,notnull,default:'medium'" fake:"oneof:low,medium,high,critical" enum:"low,medium,high,critical"`
	Status          string `json:"status" bun:"status,notnull,default:'pending'" fake:"oneof:pending,accepted,rejected,converted" enum:"pending,approved,rejected,converted"`
	ConvertedTaskID string `json:"converted_task_id" bun:"converted_task_id" fake:"-"`

	// Relations
	Project       *Project `json:"project,omitempty" bun:"-" yolo:"rel:belongs_to,fk:project_id"`
	ConvertedTask *Task    `json:"converted_task,omitempty" bun:"-" yolo:"rel:belongs_to,fk:converted_task_id"`
}

func (Suggestion) TableName() string  { return "factory_suggestions" }
func (Suggestion) EntityName() string { return "Suggestion" }
