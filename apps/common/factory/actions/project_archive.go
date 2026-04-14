package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
)

// ProjectArchive archives an active project.
//
// @policy CanArchiveProjectPolicy
func ProjectArchive(ctx context.Context, actx *action.Context) error {
	_, err := sm.Project.Archive(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("project already archived")
	}
	if err != nil {
		return fmt.Errorf("archive-project: %w", err)
	}
	return nil
}
