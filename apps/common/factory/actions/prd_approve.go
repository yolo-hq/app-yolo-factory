package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApprovePRDData declares the entity fields this action reads.
// The Project nested struct uses the PRD's belongs_to relation to load
// auto_start in the same query — no manual ReadRepo call needed.
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
	if prd.Project.AutoStart {
		jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: prd.ID})
	}

	return nil
}
