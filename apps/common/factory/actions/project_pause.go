package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// PauseProjectAction pauses an active project.
type PauseProjectAction struct {
	action.RequirePolicy[policies.CanPauseProjectPolicy]
	action.NoInput
}

func (a *PauseProjectAction) Description() string { return "Pause an active project" }

func (a *PauseProjectAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := repos.Project.UpdateEntity(ctx, actx, write.Set{
		fields.Project.Status.Value(string(enums.ProjectStatusPaused)),
	})
	return err
}
