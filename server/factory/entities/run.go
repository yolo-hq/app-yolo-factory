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
}

func (Run) TableName() string  { return "runs" }
func (Run) EntityName() string { return "Run" }

func (Run) Relations() []entity.Relation {
	return []entity.Relation{
		{Name: "Task", Type: entity.RelationManyToOne, Table: "tasks", ForeignKey: "task_id", ReferenceKey: "id"},
		{Name: "Repo", Type: entity.RelationManyToOne, Table: "repos", ForeignKey: "repo_id", ReferenceKey: "id"},
		{Name: "Questions", Type: entity.RelationOneToMany, Table: "questions", ForeignKey: "run_id", ReferenceKey: "id"},
	}
}
