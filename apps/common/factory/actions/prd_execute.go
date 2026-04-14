package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
)

// PRDExecute kicks off PRD planning by enqueuing a PlanPRDJob.
//
// @policy CanExecutePRDPolicy
func PRDExecute(ctx context.Context, actx *action.Context) error {
	_, err := sm.PRD.Execute(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("PRD is not in draft or approved state")
	}
	if err != nil {
		return fmt.Errorf("execute-prd: %w", err)
	}

	jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: actx.EntityID})
	return nil
}
