package entities

import (
	"time"

	"github.com/yolo-hq/yolo/core/entity"
)

type PRD struct {
	entity.BaseEntity
	ProjectID          string     `json:"projectId" bun:"project_id,notnull" fake:"rel:Project"`
	Title              string     `json:"title" bun:"title,notnull" fake:"sentence:6"`
	Status             string     `json:"status" bun:"status,notnull,default:'draft'" fake:"oneof:draft,approved,in_progress,completed,failed"`
	Source             string     `json:"source" bun:"source,notnull,default:'manual'" fake:"oneof:manual,suggestion,issue"`
	CreatedBy          string     `json:"createdBy" bun:"created_by,notnull,default:'human'" fake:"oneof:human,agent"`
	Body               string     `json:"body" bun:"body,notnull" fake:"sentence:20"`
	AcceptanceCriteria string     `json:"acceptanceCriteria" bun:"acceptance_criteria,notnull" fake:"sentence:15"`
	DesignDecisions    string     `json:"designDecisions" bun:"design_decisions,default:'[]'" fake:"-"`
	TotalTasks         int        `json:"totalTasks" bun:"total_tasks,default:0" fake:"int:0,10"`
	CompletedTasks     int        `json:"completedTasks" bun:"completed_tasks,default:0" fake:"int:0,5"`
	FailedTasks        int        `json:"failedTasks" bun:"failed_tasks,default:0" fake:"int:0,3"`
	TotalCostUSD       float64    `json:"totalCostUsd" bun:"total_cost_usd,default:0" fake:"float:0,50"`
	ApprovedAt         *time.Time `json:"approvedAt" bun:"approved_at" fake:"-"`
	CompletedAt        *time.Time `json:"completedAt" bun:"completed_at" fake:"-"`

	// Relations
	Project *Project `json:"project,omitempty" bun:"-" yolo:"rel:belongs_to,fk:project_id"`
	Tasks   []Task   `json:"tasks,omitempty" bun:"-" yolo:"rel:has_many,fk:prd_id"`
}

func (PRD) TableName() string  { return "factory_prds" }
func (PRD) EntityName() string { return "PRD" }
