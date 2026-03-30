package entities

import "github.com/yolo-hq/yolo/core/entity"

type Question struct {
	entity.BaseEntity
	TaskID     string `json:"taskId" bun:"task_id,notnull" fake:"rel:Task"`
	RunID      string `json:"runId" bun:"run_id,notnull" fake:"rel:Run"`
	RepoID     string `json:"repoId" bun:"repo_id,notnull" fake:"rel:Repo"`
	Status     string `json:"status" bun:"status,notnull,default:'open'" fake:"oneof:open,resolved,dismissed"`
	Context    string `json:"context" bun:"context" fake:"sentence:15"`
	Tried      string `json:"tried" bun:"tried" fake:"sentence:10"`
	Body       string `json:"body" bun:"body,notnull" fake:"sentence:20"`
	Resolution string `json:"resolution" bun:"resolution" fake:"-"`

	// Relations
	Task *Task `json:"task,omitempty" bun:"-" yolo:"rel:belongs_to,fk:task_id"`
	Run  *Run  `json:"run,omitempty" bun:"-" yolo:"rel:belongs_to,fk:run_id"`
	Repo *Repo `json:"repo,omitempty" bun:"-" yolo:"rel:belongs_to,fk:repo_id"`
}

func (Question) TableName() string  { return "questions" }
func (Question) EntityName() string { return "Question" }
