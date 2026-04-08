package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// PauseProjectAction pauses an active project.
type PauseProjectAction struct {
	action.NoInput
	action.RequirePolicy[policies.CanPauseProjectPolicy]
}

func (a *PauseProjectAction) Description() string { return "Pause an active project" }

func (a *PauseProjectAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.ExecUpdate[entities.Project](ctx, actx, write.Set{
		write.NewField[string]("status").Value(entities.ProjectPaused),
	}); r != nil {
		return *r
	}
	actx.Resolve("Project", actx.EntityID)
	return action.OK()
}
