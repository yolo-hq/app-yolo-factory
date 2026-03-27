package entities

import (
	"time"

	"github.com/yolo-hq/yolo/core/entity"
)

type Run struct {
	entity.BaseEntity
	TaskID      string     `json:"taskId" bun:"task_id,notnull"`
	RepoID      string     `json:"repoId" bun:"repo_id,notnull"`
	Agent       string     `json:"agent" bun:"agent,notnull,default:'claude-cli'"`
	Model       string     `json:"model" bun:"model,notnull"`
	Status      string     `json:"status" bun:"status,notnull,default:'running'"`
	Cost        float64    `json:"cost" bun:"cost,notnull,default:0"`
	Duration    int        `json:"duration" bun:"duration,notnull,default:0"`
	LogURL      string     `json:"logUrl" bun:"log_url"`
	Error       string     `json:"error" bun:"error"`
	CommitHash  string     `json:"commitHash" bun:"commit_hash"`
	StartedAt   time.Time  `json:"startedAt" bun:"started_at,notnull,default:current_timestamp"`
	CompletedAt *time.Time `json:"completedAt" bun:"completed_at"`

	// Relations
	Task      *Task      `json:"task,omitempty" yolo:"rel:belongs_to,fk:task_id"`
	Repo      *Repo      `json:"repo,omitempty" yolo:"rel:belongs_to,fk:repo_id"`
	Questions []Question `json:"questions,omitempty" yolo:"rel:has_many,fk:run_id"`
}

func (Run) TableName() string  { return "runs" }
func (Run) EntityName() string { return "Run" }
