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
}

func (a *ResumeProjectAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{&policies.ProjectMustBePaused{}}
}

func (a *ResumeProjectAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	_, err := action.Write[entities.Project](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{write.NewField[string]("status").Value(entities.ProjectActive)},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Project", actx.EntityID)
	return action.OK()
}
