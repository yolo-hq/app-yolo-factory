package actions

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CreateRunInput struct {
	TaskID string `json:"taskId" validate:"required"`
	RepoID string `json:"repoId" validate:"required"`
	Agent  string `json:"agent"`
	Model  string `json:"model" validate:"required"`
}

type CreateRunAction struct {
	action.BaseCreate[entities.Run, CreateRunInput]
	Repo entity.WriteRepository[entities.Run]
}

func (a *CreateRunAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}
