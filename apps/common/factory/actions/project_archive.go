package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ArchiveProjectAction archives an active project.
type ArchiveProjectAction struct {
	action.RequirePolicy[policies.CanArchiveProjectPolicy]
	action.NoInput
}

func (a *ArchiveProjectAction) Description() string { return "Archive a project" }

func (a *ArchiveProjectAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := sm.Project.Archive(ctx, actx, actx.EntityID, nil)
	return err
}
