package actions

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CreateRepoInput struct {
	Name         string `json:"name" validate:"required"`
	URL          string `json:"url" validate:"required"`
	LocalPath    string `json:"localPath"`
	TargetBranch string `json:"targetBranch"`
	DefaultModel string `json:"defaultModel"`
}

type CreateRepoAction struct {
	action.BaseCreate[entities.Repo, CreateRepoInput]
	Repo entity.WriteRepository[entities.Repo]
}

func (a *CreateRepoAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}
