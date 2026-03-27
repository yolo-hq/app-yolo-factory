package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CancelTaskAction struct {
	action.NoInput
}

func (a *CancelTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	_, err := action.WriteRepo[entities.Task](actx).Update(ctx).
		WhereID(actx.EntityID).
		Set("status", "cancelled").
		Exec(ctx)
	if err != nil {
		return action.Failure("cancel failed: " + err.Error())
	}

	return action.OK()
}
