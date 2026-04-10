package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
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
	_, err := action.Write[entities.Project](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{fields.Project.Status.Value(string(enums.ProjectStatusPaused))},
	})
	if err != nil {
		return action.Failure(err.Error())
	}
	actx.Resolve("Project", actx.EntityID)
	return action.OK()
}
