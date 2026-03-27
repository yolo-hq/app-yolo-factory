package entities

import "github.com/yolo-hq/yolo/core/entity"

type Task struct {
	entity.BaseEntity
	RepoID      string  `json:"repoId" bun:"repo_id,notnull"`
	Title       string  `json:"title" bun:"title,notnull"`
	Body        string  `json:"body" bun:"body"`
	Type        string  `json:"type" bun:"type,notnull,default:'auto'"`
	Status      string  `json:"status" bun:"status,notnull,default:'queued'"`
	Priority    int     `json:"priority" bun:"priority,notnull,default:3"`
	Model       string  `json:"model" bun:"model"`
	Labels      string  `json:"labels" bun:"labels,default:'[]'"`
	ParentID    string  `json:"parentId" bun:"parent_id"`
	DependsOn   string  `json:"dependsOn" bun:"depends_on,default:'[]'"`
	Cost        float64 `json:"cost" bun:"cost,notnull,default:0"`
	RunCount    int     `json:"runCount" bun:"run_count,notnull,default:0"`
	MaxRetries  int     `json:"maxRetries" bun:"max_retries,notnull,default:3"`
	TimeoutSecs int     `json:"timeoutSecs" bun:"timeout_secs,notnull,default:600"`

	// Relations
	Repo      *Repo      `json:"repo,omitempty" bun:"-" yolo:"rel:belongs_to,fk:repo_id"`
	Runs      []Run      `json:"runs,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Questions []Question `json:"questions,omitempty" bun:"-" yolo:"rel:has_many,fk:task_id"`
	Parent    *Task      `json:"parent,omitempty" bun:"-" yolo:"rel:belongs_to,fk:parent_id"`
	Children  []Task     `json:"children,omitempty" bun:"-" yolo:"rel:has_many,fk:parent_id"`
}

func (Task) TableName() string  { return "tasks" }
func (Task) EntityName() string { return "Task" }
