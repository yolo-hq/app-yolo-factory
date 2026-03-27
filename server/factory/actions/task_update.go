package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type UpdateTaskAction struct {
	action.TypedInput[inputs.UpdateTaskInput]
	Repo entity.WriteRepository[entities.Task]
}


func (a *UpdateTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.RequireEntityID(actx, "Task"); r != nil {
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
		return action.NotFound("Task", actx.EntityID)
	}

	return action.Success(updated, "updated")
}
