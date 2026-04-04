package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// CancelTaskAction cancels a task.
type CancelTaskAction struct {
	action.NoInput
}

func (a *CancelTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	_, r := action.FindOrFail[entities.Task](ctx, action.ReadRepo[entities.Task](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	_, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{write.NewField[string]("status").Value("cancelled")},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Task", actx.EntityID)
	return action.OK()
}
