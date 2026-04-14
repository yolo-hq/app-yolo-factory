package entities

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Task struct {
	bun.BaseModel `bun:"table:factory_tasks"`
	entity.BaseEntity
	PrdID              string     `json:"prd_id" bun:"prd_id" fake:"rel:PRD"`
	ProjectID          string     `json:"project_id" bun:"project_id" fake:"rel:Project"`
	Title              string     `json:"title" bun:"title,notnull" fake:"sentence:6"`
	Status             string     `json:"status" bun:"status,notnull,default:'queued'" fake:"oneof:queued,running,completed,failed,cancelled" enum:"queued,blocked,running,reviewing,done,failed,cancelled"`
	Spec               string     `json:"spec" bun:"spec,notnull" fake:"sentence:20"`
	AcceptanceCriteria string     `json:"acceptance_criteria" bun:"acceptance_criteria,notnull" fake:"sentence:15"`
	Branch             string     `json:"branch" bun:"branch,notnull" fake:"word"`
	Model              string     `json:"model" bun:"model,default:''" fake:"oneof:sonnet,opus,haiku"`
	Sequence           int        `json:"sequence" bun:"sequence,notnull" fake:"int:1,10"`
	DependsOn          string     `json:"depends_on" bun:"depends_on,default:'[]'" fake:"-"`
	RunCount           int        `json:"run_count" bun:"run_count,default:0" fake:"int:0,5"`
	MaxRetries         int        `json:"max_retries" bun:"max_retries,default:3" fake:"int:1,5"`
	CostUSD            float64    `json:"cost_usd" bun:"cost_usd,default:0" fake:"float:0,25"`
	Summary            string     `json:"summary" bun:"summary" fake:"sentence:10"`
	CommitHash         string     `json:"commit_hash" bun:"commit_hash" fake:"-"`
	StartedAt          *time.Time `json:"started_at" bun:"started_at" fake:"-"`
	CompletedAt        *time.Time `json:"completed_at" bun:"completed_at" fake:"-"`

	// Relations
	PRD       *PRD       `json:"prd,omitempty" bun:"-" yolo:"rel:belongs_to,fk:prd_id"`
	Project   *Project   `json:"project,omitempty" bun:"-" yolo:"rel:belongs_to,fk:project_id"`
	Runs      []Run      `json:"runs,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Reviews   []Review   `json:"reviews,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Questions []Question `json:"questions,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
}

func (Task) TableName() string  { return "factory_tasks" }
func (Task) EntityName() string { return "Task" }
