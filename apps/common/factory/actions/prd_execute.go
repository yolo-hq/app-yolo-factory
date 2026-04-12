package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
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

	// Transition to planning with conditional where for race-safety.
	_, err := repos.PRD.UpdateWhere(ctx, actx, actx.EntityID,
		[]entity.FilterCondition{
			{Field: "status", Operator: entity.OpIn, Value: []string{
				string(enums.PRDStatusDraft),
				string(enums.PRDStatusApproved),
			}},
		},
		write.Set{
			fields.PRD.Status.Value(string(enums.PRDStatusPlanning)),
		},
	)
	if err != nil {
		return err
	}

	// Defer planning job until after the tx commits.
	jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: prd.ID})

	return nil
}
