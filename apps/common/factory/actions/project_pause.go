package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// PauseProjectAction pauses an active project.
type PauseProjectAction struct {
	action.RequirePolicy[policies.CanPauseProjectPolicy]
	action.NoInput
}

func (a *PauseProjectAction) Description() string { return "Pause an active project" }

func (a *PauseProjectAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := sm.Project.Pause(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("project is not active")
	}
	if err != nil {
		return fmt.Errorf("pause-project: %w", err)
	}
	return nil
}
