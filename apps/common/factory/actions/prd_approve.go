package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
)

// ApprovePRDData reads only the PRD fields needed for this action.
type ApprovePRDData struct {
	projection.For[entities.PRD]

	ID        string `field:"id"`
	ProjectID string `field:"project_id"`
}

// projectAutoStart reads only the auto_start field from a project.
type projectAutoStart struct {
	projection.For[entities.Project]
	AutoStart bool `field:"auto_start"`
}

// PRDApprove approves a draft PRD and optionally triggers planning.
//
// @policy CanApprovePRDPolicy
func PRDApprove(ctx context.Context, actx *action.Context) error {
	prd, err := projection.Load[ApprovePRDData](ctx, actx.EntityID)
	if err != nil {
		return fmt.Errorf("approve-prd: load prd: %w", err)
	}

	now := time.Now()
	_, err = sm.PRD.Approve(ctx, actx, actx.EntityID, write.Set{
		fields.PRD.ApprovedAt.Value(&now),
	})
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("PRD is not in draft state")
	}
	if err != nil {
		return fmt.Errorf("approve-prd: %w", err)
	}

	project, err := read.FindOne[projectAutoStart](ctx, prd.ProjectID)
	if err != nil {
		return fmt.Errorf("approve-prd: load project: %w", err)
	}

	if project.AutoStart {
		jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: prd.ID})
	}

	return nil
}
