package entities

import "github.com/yolo-hq/yolo/core/entity"

type Task struct {
	entity.BaseEntity
	RepoID      string  `json:"repoId" bun:"repo_id,notnull" fake:"rel:Repo,5x"`
	Title       string  `json:"title" bun:"title,notnull" fake:"sentence:6"`
	Body        string  `json:"body" bun:"body" fake:"sentence:20"`
	Type        string  `json:"type" bun:"type,notnull,default:'auto'" fake:"oneof:auto,manual,review"`
	Status      string  `json:"status" bun:"status,notnull,default:'queued'" fake:"oneof:queued,running,completed,failed,cancelled"`
	Priority    int     `json:"priority" bun:"priority,notnull,default:3" fake:"int:1,5"`
	Model       string  `json:"model" bun:"model" fake:"oneof:sonnet,opus,haiku"`
	Labels      string  `json:"labels" bun:"labels,default:'[]'" fake:"-"`
	ParentID    string  `json:"parentId" bun:"parent_id" fake:"-"`
	DependsOn   string  `json:"dependsOn" bun:"depends_on,default:'[]'" fake:"-"`
	Cost        float64 `json:"cost" bun:"cost,notnull,default:0" fake:"float:0,50"`
	RunCount    int     `json:"runCount" bun:"run_count,notnull,default:0" fake:"int:0,5"`
	MaxRetries  int     `json:"maxRetries" bun:"max_retries,notnull,default:3" fake:"int:1,5"`
	TimeoutSecs int     `json:"timeoutSecs" bun:"timeout_secs,notnull,default:600" fake:"int:60,1800"`

	// Relations
	Repo      *Repo      `json:"repo,omitempty" bun:"-" yolo:"rel:belongs_to,fk:repo_id"`
	Runs      []Run      `json:"runs,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Questions []Question `json:"questions,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Parent    *Task      `json:"parent,omitempty" bun:"-" yolo:"rel:belongs_to,fk:parent_id"`
	Children  []Task     `json:"children,omitempty" bun:"-" yolo:"rel:has_many,fk:parent_id"`
}

func (Task) TableName() string  { return "tasks" }
func (Task) EntityName() string { return "Task" }
