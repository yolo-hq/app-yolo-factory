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

}

func (Task) TableName() string  { return "tasks" }
func (Task) EntityName() string { return "Task" }
