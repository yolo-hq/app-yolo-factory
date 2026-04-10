package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApprovePRDData declares the entity fields this action reads.
type ApprovePRDData struct {
	ID        string `field:"id"`
	ProjectID string `field:"project_id"`
}

// ApprovePRDAction approves a draft PRD and optionally triggers planning.
type ApprovePRDAction struct {
	action.NoInput
	action.RequirePolicy[policies.CanApprovePRDPolicy]
	action.TypedData[ApprovePRDData]
}

func (a *ApprovePRDAction) Description() string { return "Approve a draft PRD" }

func (a *ApprovePRDAction) Execute(ctx context.Context, actx *action.Context) error {
	prd := a.Data(actx)

	now := time.Now()
	_, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Where: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: string(enums.PRDStatusDraft)},
		},
		Set: write.Set{
			fields.PRD.Status.Value(string(enums.PRDStatusApproved)),
			fields.PRD.ApprovedAt.Value(&now),
		},
	})
	if err != nil {
		return err
	}

	// auto_start: if project has AutoStart, defer PlanPRDJob until after commit.
	project, err := action.ReadRepo[entities.Project](actx).FindOne(ctx, entity.FindOneOptions{ID: prd.ProjectID})
	if err == nil && project != nil && project.AutoStart {
		jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: prd.ID})
	}

	return nil
}
