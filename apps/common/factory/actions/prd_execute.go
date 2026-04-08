package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ExecutePRDAction kicks off PRD planning by enqueuing a PlanPRDJob.
type ExecutePRDAction struct {
	action.NoInput
	JobClient *jobs.Client
	PlanJob   jobs.Handler
}

func (a *ExecutePRDAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{&policies.PRDMustBeDraftOrApproved{}}
}

func (a *ExecutePRDAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	prd, r := action.FindOrFail[entities.PRD](ctx, action.ReadRepo[entities.PRD](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	// Transition to planning.
	_, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			write.NewField[string]("status").Value(entities.PRDPlanning),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	// Enqueue planning job.
	_, err = a.JobClient.Dispatch(ctx, a.PlanJob, map[string]string{
		"prd_id": prd.ID,
	})
	if err != nil {
		return action.Failure("failed to enqueue planning job: " + err.Error())
	}

	actx.Resolve("PRD", actx.EntityID)
	return action.OK(map[string]string{"status": entities.PRDPlanning})
}
