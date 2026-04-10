package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ExecutePRDData declares the entity fields this action reads.
type ExecutePRDData struct {
	ID string `field:"id"`
}

// ExecutePRDAction kicks off PRD planning by enqueuing a PlanPRDJob.
type ExecutePRDAction struct {
	action.NoInput
	action.RequirePolicy[policies.CanExecutePRDPolicy]
	action.TypedData[ExecutePRDData]
	JobClient *jobs.Client
	PlanJob   jobs.Handler
}

func (a *ExecutePRDAction) Description() string { return "Execute a PRD by starting planning" }

func (a *ExecutePRDAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	prd := a.Data(actx)

	// Transition to planning.
	_, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			fields.PRD.Status.Value(string(enums.PRDStatusPlanning)),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	// Defer planning job until after the tx commits. If the update above
	// rolls back, the job will not be dispatched.
	actx.DeferJob(a.PlanJob, map[string]string{
		"prd_id": prd.ID,
	})

	actx.Resolve("PRD", actx.EntityID)
	return action.OK(map[string]string{"status": string(enums.PRDStatusPlanning)})
}
