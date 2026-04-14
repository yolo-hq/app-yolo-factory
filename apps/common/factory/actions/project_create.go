package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// ProjectCreate creates a new project.
func ProjectCreate(ctx context.Context, actx *action.Context, in inputs.CreateProjectInput) error {
	res, err := repos.Project.CreateFromInput(ctx, actx, in)
	if err != nil {
		return err
	}
	actx.Resolve("Project", res.ID())
	return nil
}
