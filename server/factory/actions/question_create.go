package actions

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CreateQuestionInput struct {
	TaskID  string `json:"taskId" validate:"required"`
	RunID   string `json:"runId" validate:"required"`
	RepoID  string `json:"repoId" validate:"required"`
	Context string `json:"context"`
	Tried   string `json:"tried"`
	Body    string `json:"body" validate:"required"`
}

type CreateQuestionAction struct {
	action.BaseCreate[entities.Question, CreateQuestionInput]
	Repo entity.WriteRepository[entities.Question]
}

func (a *CreateQuestionAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}
