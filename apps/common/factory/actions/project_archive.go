package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ArchiveProjectAction archives an active project.
type ArchiveProjectAction struct {
	action.RequirePolicy[policies.CanArchiveProjectPolicy]
	action.NoInput
}

func (a *ArchiveProjectAction) Description() string { return "Archive a project" }

func (a *ArchiveProjectAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := repos.Project.UpdateEntity(ctx, actx, write.Set{
		fields.Project.Status.Value(string(enums.ProjectStatusArchived)),
	})
	if err != nil {
		return fmt.Errorf("archive-project: %w", err)
	}
	return nil
}
