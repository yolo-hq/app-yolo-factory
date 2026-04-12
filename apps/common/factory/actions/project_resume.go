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

// ResumeProjectAction resumes a paused project.
type ResumeProjectAction struct {
	action.RequirePolicy[policies.CanResumeProjectPolicy]
	action.NoInput
}

func (a *ResumeProjectAction) Description() string { return "Resume a paused project" }

func (a *ResumeProjectAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := action.Write[entities.Project](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{fields.Project.Status.Value(string(enums.ProjectStatusActive))},
	})
	return err
}
