package actions

import (
	"context"

	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CancelTaskAction struct {
	action.NoInput
	Repo entity.WriteRepository[entities.Task]
}

func (a *CancelTaskAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}

func (a *CancelTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.RequireEntityID(actx, "Task"); r != nil {
		return *r
	}

	builder := a.Repo.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: actx.EntityID}).
		Set("status", "cancelled").
		Returning()

	updated, err := builder.Exec(ctx)
	if err != nil {
		return action.Failure("cancel failed: " + err.Error())
	}
	if updated == nil {
		return action.NotFound("Task", actx.EntityID)
	}

	return action.Success(updated, "cancelled")
}
