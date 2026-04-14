package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApprovePRDData reads only the PRD fields needed for this action.
type ApprovePRDData struct {
	projection.For[entities.PRD]

	ID      string `field:"id"`
	Project struct {
		AutoStart bool `field:"auto_start"`
	} `field:"project"`
}

// ApprovePRDAction approves a draft PRD and optionally triggers planning.
type ApprovePRDAction struct {
	action.RequirePolicy[policies.CanApprovePRDPolicy]
	action.NoInput
	action.Projection[ApprovePRDData]
}

func (a *ApprovePRDAction) Description() string { return "Approve a draft PRD" }

func (a *ApprovePRDAction) Execute(ctx context.Context, actx *action.Context) error {
	prd := a.Data(actx)

	now := time.Now()
	if _, err := sm.PRD.Approve(ctx, actx, actx.EntityID, write.Set{
		fields.PRD.ApprovedAt.Value(&now),
	}); err != nil {
		return err
	}

	if prd.Project.AutoStart {
		jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: prd.ID})
	}

	return nil
}
