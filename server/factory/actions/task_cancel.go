package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CancelTaskAction struct {
	action.NoInput
	Repo entity.WriteRepository[entities.Task]
}


func (a *CancelTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.RequireEntityID(actx, "Task"); r != nil {
		return *r
	}

	_, err := a.Repo.Update(ctx).
		WhereID(actx.EntityID).
		Set("status", "cancelled").
		Exec(ctx)
	if err != nil {
		return action.Failure("cancel failed: " + err.Error())
	}

	actx.Resolve("Task", actx.EntityID)
	return action.OK()
}
