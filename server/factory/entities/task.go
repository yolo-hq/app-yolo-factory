package entities

import (
	"time"

	"github.com/yolo-hq/yolo/core/entity"
)

type Task struct {
	entity.BaseEntity
	PrdID              string     `json:"prdId" bun:"prd_id" fake:"rel:PRD"`
	ProjectID          string     `json:"projectId" bun:"project_id" fake:"rel:Project"`
	Title              string     `json:"title" bun:"title,notnull" fake:"sentence:6"`
	Status             string     `json:"status" bun:"status,notnull,default:'queued'" fake:"oneof:queued,running,completed,failed,cancelled"`
	Spec               string     `json:"spec" bun:"spec,notnull" fake:"sentence:20"`
	AcceptanceCriteria string     `json:"acceptanceCriteria" bun:"acceptance_criteria,notnull" fake:"sentence:15"`
	Branch             string     `json:"branch" bun:"branch,notnull" fake:"word"`
	Model              string     `json:"model" bun:"model,default:''" fake:"oneof:sonnet,opus,haiku"`
	Sequence           int        `json:"sequence" bun:"sequence,notnull" fake:"int:1,10"`
	DependsOn          string     `json:"dependsOn" bun:"depends_on,default:'[]'" fake:"-"`
	RunCount           int        `json:"runCount" bun:"run_count,default:0" fake:"int:0,5"`
	MaxRetries         int        `json:"maxRetries" bun:"max_retries,default:3" fake:"int:1,5"`
	CostUSD            float64    `json:"costUsd" bun:"cost_usd,default:0" fake:"float:0,25"`
	Summary            string     `json:"summary" bun:"summary" fake:"sentence:10"`
	CommitHash         string     `json:"commitHash" bun:"commit_hash" fake:"-"`
	StartedAt          *time.Time `json:"startedAt" bun:"started_at" fake:"-"`
	CompletedAt        *time.Time `json:"completedAt" bun:"completed_at" fake:"-"`

	// Relations
	PRD       *PRD       `json:"prd,omitempty" bun:"-" yolo:"rel:belongs_to,fk:prd_id"`
	Project   *Project   `json:"project,omitempty" bun:"-" yolo:"rel:belongs_to,fk:project_id"`
	Runs      []Run      `json:"runs,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Reviews   []Review   `json:"reviews,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Questions []Question `json:"questions,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
}

func (Task) TableName() string  { return "factory_tasks" }
func (Task) EntityName() string { return "Task" }
