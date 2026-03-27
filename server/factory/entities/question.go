package entities

import "github.com/yolo-hq/yolo/core/entity"

type Question struct {
	entity.BaseEntity
	TaskID     string `json:"taskId" bun:"task_id,notnull"`
	RunID      string `json:"runId" bun:"run_id,notnull"`
	RepoID     string `json:"repoId" bun:"repo_id,notnull"`
	Status     string `json:"status" bun:"status,notnull,default:'open'"`
	Context    string `json:"context" bun:"context"`
	Tried      string `json:"tried" bun:"tried"`
	Body       string `json:"body" bun:"body,notnull"`
	Resolution string `json:"resolution" bun:"resolution"`

	// Relations
	Task *Task `json:"task,omitempty" bun:"rel:belongs_to,join:task_id=id"`
	Run  *Run  `json:"run,omitempty" bun:"rel:belongs_to,join:run_id=id"`
	Repo *Repo `json:"repo,omitempty" bun:"rel:belongs_to,join:repo_id=id"`
}

func (Question) TableName() string  { return "questions" }
func (Question) EntityName() string { return "Question" }
