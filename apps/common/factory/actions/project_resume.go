package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ResumeProjectAction resumes a paused project.
type ResumeProjectAction struct {
	action.RequirePolicy[policies.CanResumeProjectPolicy]
	action.NoInput
}

func (a *ResumeProjectAction) Description() string { return "Resume a paused project" }

func (a *ResumeProjectAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := sm.Project.Resume(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("project is not paused")
	}
	if err != nil {
		return fmt.Errorf("resume-project: %w", err)
	}
	return nil
}
