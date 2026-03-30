package entities

import (
	"time"

	"github.com/yolo-hq/yolo/core/entity"
)

type Run struct {
	entity.BaseEntity
	TaskID      string     `json:"taskId" bun:"task_id,notnull" fake:"rel:Task,2x"`
	RepoID      string     `json:"repoId" bun:"repo_id,notnull" fake:"rel:Repo"`
	Agent       string     `json:"agent" bun:"agent,notnull,default:'claude-cli'" fake:"oneof:claude-cli,claude-code,custom"`
	Model       string     `json:"model" bun:"model,notnull" fake:"oneof:sonnet,opus,haiku"`
	Status      string     `json:"status" bun:"status,notnull,default:'running'" fake:"oneof:running,completed,failed,cancelled"`
	Cost        float64    `json:"cost" bun:"cost,notnull,default:0" fake:"float:0,25"`
	Duration    int        `json:"duration" bun:"duration,notnull,default:0" fake:"int:10,3600"`
	LogURL      string     `json:"logUrl" bun:"log_url" fake:"url"`
	Error       string     `json:"error" bun:"error" fake:"-"`
	CommitHash  string     `json:"commitHash" bun:"commit_hash" fake:"uuid"`
	StartedAt   time.Time  `json:"startedAt" bun:"started_at,notnull,default:current_timestamp"`
	CompletedAt *time.Time `json:"completedAt" bun:"completed_at" fake:"-"`

	// Relations
	Task      *Task      `json:"task,omitempty" bun:"-" yolo:"rel:belongs_to,fk:task_id"`
	Repo      *Repo      `json:"repo,omitempty" bun:"-" yolo:"rel:belongs_to,fk:repo_id"`
	Questions []Question `json:"questions,omitempty" bun:"-" yolo:"rel:has_many,fk:run_id"`
}

func (Run) TableName() string  { return "runs" }
func (Run) EntityName() string { return "Run" }
