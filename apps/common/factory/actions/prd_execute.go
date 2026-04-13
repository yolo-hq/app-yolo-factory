package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ExecutePRDData declares the entity fields this action reads.
type ExecutePRDData struct {
	projection.For[entities.PRD]

	ID string `field:"id"`
}

// ExecutePRDAction kicks off PRD planning by enqueuing a PlanPRDJob.
type ExecutePRDAction struct {
	action.RequirePolicy[policies.CanExecutePRDPolicy]
	action.NoInput
	action.Projection[ExecutePRDData]
}

func (a *ExecutePRDAction) Description() string { return "Execute a PRD by starting planning" }

func (a *ExecutePRDAction) Execute(ctx context.Context, actx *action.Context) error {
	prd := a.Data(actx)

	// Transition to planning. SM enforces draft|approved → planning.
	_, err := sm.PRD.Execute(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("PRD is not in draft or approved state")
	}
	if err != nil {
		return fmt.Errorf("execute-prd: %w", err)
	}

	// Defer planning job until after the tx commits.
	jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: prd.ID})

	return nil
}
