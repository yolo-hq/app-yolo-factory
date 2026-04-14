package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
)

// ProjectPause pauses an active project.
//
// @policy CanPauseProjectPolicy
func ProjectPause(ctx context.Context, actx *action.Context) error {
	_, err := sm.Project.Pause(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("project is not active")
	}
	if err != nil {
		return fmt.Errorf("pause-project: %w", err)
	}
	return nil
}
