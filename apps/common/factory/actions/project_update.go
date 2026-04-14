package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// ProjectUpdate updates an existing project.
func ProjectUpdate(ctx context.Context, actx *action.Context, in inputs.UpdateProjectInput) error {
	if _, err := repos.Project.UpdateFromInput(ctx, actx, actx.EntityID, in); err != nil {
		return err
	}
	actx.Resolve("Project", actx.EntityID)
	return nil
}
