package entities

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type PRD struct {
	bun.BaseModel `bun:"table:factory_prds"`
	entity.BaseEntity
	ProjectID          string     `json:"project_id" bun:"project_id,notnull" fake:"rel:Project"`
	Title              string     `json:"title" bun:"title,notnull" fake:"sentence:6"`
	Status             string     `json:"status" bun:"status,notnull,default:'draft'" fake:"oneof:draft,approved,in_progress,completed,failed" enum:"draft,approved,planning,in_progress,completed,failed"`
	Source             string     `json:"source" bun:"source,notnull,default:'manual'" fake:"oneof:manual,suggestion,issue" enum:"manual,grill_me,factory_generated,imported"`
	CreatedBy          string     `json:"created_by" bun:"created_by,notnull,default:'human'" fake:"oneof:human,agent"`
	Body               string     `json:"body" bun:"body,notnull" fake:"sentence:20"`
	AcceptanceCriteria string     `json:"acceptance_criteria" bun:"acceptance_criteria,notnull" fake:"sentence:15"`
	DesignDecisions    string     `json:"design_decisions" bun:"design_decisions,default:'[]'" fake:"-"`
	FailedTasks        int        `json:"failed_tasks" bun:"failed_tasks,default:0" fake:"int:0,3"`
	ApprovedAt         *time.Time `json:"approved_at" bun:"approved_at" fake:"-"`
	CompletedAt        *time.Time `json:"completed_at" bun:"completed_at" fake:"-"`

	// Virtual computed fields (not stored — derived from tasks relation).
	CompletedTasks int     `bun:"-" json:"completed_tasks" aggregate:"count" rel:"tasks" filter:"status=done"`
	TotalTasks     int     `bun:"-" json:"total_tasks" aggregate:"count" rel:"tasks"`
	TotalCostUSD   float64 `bun:"-" json:"total_cost_usd" aggregate:"sum:cost_usd" rel:"tasks"`
	Progress       int     `bun:"-" json:"progress" computed:"transform"`

	// Relations
	Project *Project `json:"project,omitempty" bun:"-" yolo:"rel:belongs_to,fk:project_id"`
	Tasks   []Task   `json:"tasks,omitempty" bun:"-" yolo:"rel:has_many,fk:prd_id"`
}

func (PRD) TableName() string  { return "factory_prds" }
func (PRD) EntityName() string { return "PRD" }

// ComputeProgress returns completion percentage (0–100).
func (p *PRD) ComputeProgress() int {
	if p.TotalTasks == 0 {
		return 0
	}
	return p.CompletedTasks * 100 / p.TotalTasks
}
