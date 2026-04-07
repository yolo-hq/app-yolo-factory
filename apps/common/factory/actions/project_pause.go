package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// PauseProjectAction pauses an active project.
type PauseProjectAction struct {
	action.NoInput
}

func (a *PauseProjectAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	project, r := action.FindOrFail[entities.Project](ctx, action.ReadRepo[entities.Project](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	if project.Status != entities.ProjectActive {
		return action.Failure("project must be active to pause")
	}

	_, err := action.Write[entities.Project](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{write.NewField[string]("status").Value(entities.ProjectPaused)},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Project", actx.EntityID)
	return action.OK()
}
