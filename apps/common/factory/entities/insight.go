package entities

import (
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Insight struct {
	bun.BaseModel `bun:"table:factory_insights"`
	entity.BaseEntity
	ProjectID      string `json:"projectId" bun:"project_id" fake:"rel:Project"`
	Category       string `json:"category" bun:"category,notnull" fake:"oneof:retry_rate,cost_optimization,model_selection,spec_quality,gate_effectiveness,workflow_optimization" enum:"retry_rate,cost_optimization,model_selection,spec_quality,gate_effectiveness,workflow_optimization"`
	Title          string `json:"title" bun:"title,notnull" fake:"sentence:6"`
	Body           string `json:"body" bun:"body,notnull" fake:"sentence:20"`
	MetricData     string `json:"metricData" bun:"metric_data,notnull,default:'{}'" fake:"-"`
	Recommendation string `json:"recommendation" bun:"recommendation,notnull,default:''" fake:"sentence:10"`
	Priority       string `json:"priority" bun:"priority,notnull,default:'medium'" fake:"oneof:low,medium,high,critical"`
	Status         string `json:"status" bun:"status,notnull,default:'pending'" fake:"oneof:pending,acknowledged,applied,dismissed" enum:"pending,acknowledged,applied,dismissed"`

	// Relations
	Project *Project `json:"project,omitempty" bun:"-" yolo:"rel:belongs_to,fk:project_id"`
}

func (Insight) TableName() string  { return "factory_insights" }
func (Insight) EntityName() string { return "Insight" }
