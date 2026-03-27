package entities

import (
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/registry"
)

func init() {
	registry.RegisterGlobalEntity(Repo{})
}

type Repo struct {
	entity.BaseEntity
	Name          string `json:"name" bun:"name,notnull,unique"`
	URL           string `json:"url" bun:"url,notnull"`
	LocalPath     string `json:"localPath" bun:"local_path"`
	TargetBranch  string `json:"targetBranch" bun:"target_branch,notnull,default:'main'"`
	DefaultModel  string `json:"defaultModel" bun:"default_model,notnull,default:'sonnet'"`
	FeedbackLoops string `json:"feedbackLoops" bun:"feedback_loops,default:'[]'"`
	Active        bool   `json:"active" bun:"active,notnull,default:true"`
}

func (Repo) TableName() string  { return "repos" }
func (Repo) EntityName() string { return "Repo" }
