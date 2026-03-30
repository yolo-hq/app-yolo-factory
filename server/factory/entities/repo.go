package entities

import "github.com/yolo-hq/yolo/core/entity"

type Repo struct {
	entity.BaseEntity
	Name          string `json:"name" bun:"name,notnull,unique" fake:"name"`
	URL           string `json:"url" bun:"url,notnull" fake:"url"`
	LocalPath     string `json:"localPath" bun:"local_path" fake:"-"`
	TargetBranch  string `json:"targetBranch" bun:"target_branch,notnull,default:'main'" fake:"oneof:main,develop,staging"`
	DefaultModel  string `json:"defaultModel" bun:"default_model,notnull,default:'sonnet'" fake:"oneof:sonnet,opus,haiku"`
	FeedbackLoops string `json:"feedbackLoops" bun:"feedback_loops,default:'[]'" fake:"-"`
	Active        bool   `json:"active" bun:"active,notnull,default:true" fake:"bool"`

	// Relations
	Tasks []Task `json:"tasks,omitempty" bun:"-" yolo:"rel:has_many,fk:repo_id"`
	Runs  []Run  `json:"runs,omitempty" bun:"-" yolo:"rel:has_many,fk:repo_id"`
}

func (Repo) TableName() string  { return "repos" }
func (Repo) EntityName() string { return "Repo" }
