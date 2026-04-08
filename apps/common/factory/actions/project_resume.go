package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ResumeProjectAction resumes a paused project.
type ResumeProjectAction struct {
	action.NoInput
	action.RequirePolicy[policies.CanResumeProjectPolicy]
}

func (a *ResumeProjectAction) Description() string { return "Resume a paused project" }

func (a *ResumeProjectAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.ExecUpdate[entities.Project](ctx, actx, write.Set{
		write.NewField[string]("status").Value(entities.ProjectActive),
	}); r != nil {
		return *r
	}
	actx.Resolve("Project", actx.EntityID)
	return action.OK()
}
