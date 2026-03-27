package actions

import (
	"context"

	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type UpdateRepoInput struct {
	Name         *string `json:"name"`
	URL          *string `json:"url"`
	LocalPath    *string `json:"localPath"`
	TargetBranch *string `json:"targetBranch"`
	DefaultModel *string `json:"defaultModel"`
	Active       *bool   `json:"active"`
}

type UpdateRepoAction struct {
	action.TypedInput[UpdateRepoInput]
	Repo entity.WriteRepository[entities.Repo]
}

func (a *UpdateRepoAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}

func (a *UpdateRepoAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.RequireEntityID(actx, "Repo"); r != nil {
		return *r
	}
	input, r := a.Input(actx)
	if r != nil {
		return *r
	}

	builder := a.Repo.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: actx.EntityID}).
		SetFromInput(input).
		Returning()

	updated, err := builder.Exec(ctx)
	if err != nil {
		return action.Failure("update failed: " + err.Error())
	}
	if updated == nil {
		return action.NotFound("Repo", actx.EntityID)
	}

	return action.Success(updated, "updated")
}
