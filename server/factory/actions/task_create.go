package actions

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CreateTaskInput struct {
	RepoID      string `json:"repoId" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Body        string `json:"body"`
	Type        string `json:"type"`
	Priority    int    `json:"priority"`
	Model       string `json:"model"`
	Labels      string `json:"labels"`
	ParentID    string `json:"parentId"`
	DependsOn   string `json:"dependsOn"`
	MaxRetries  int    `json:"maxRetries"`
	TimeoutSecs int    `json:"timeoutSecs"`
}

type CreateTaskAction struct {
	action.BaseCreate[entities.Task, CreateTaskInput]
	Repo entity.WriteRepository[entities.Task]
}

func (a *CreateTaskAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}
