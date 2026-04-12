package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApprovePRDData reads only the PRD fields needed for this action.
// NOTE: The Project relation is NOT loaded here because LoadEntityWithIncludes
// cannot resolve yolo-tag belongs_to relations (bun:"-" fields). The project
// is loaded manually in Execute via read.FindOne.
// TODO(framework): LoadEntityWithIncludes should support yolo to-one relations
// the same way LoadEntityWithSpec does (see entity_loader.go yoloToOneIncludes).
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
	_, err := repos.PRD.UpdateWhere(ctx, actx, actx.EntityID,
		[]entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: string(enums.PRDStatusDraft)},
		},
		write.Set{
			fields.PRD.Status.Value(string(enums.PRDStatusApproved)),
			fields.PRD.ApprovedAt.Value(&now),
		},
	)
	if err != nil {
		return fmt.Errorf("approve-prd: %w", err)
	}

	// Load project to check auto_start. Manual load is required because the
	// framework's LoadEntityWithIncludes path cannot resolve yolo belongs_to
	// relations (bun:"-" fields) — only LoadEntityWithSpec can.
	project, err := read.FindOne[projectAutoStart](ctx, prd.ProjectID)
	if err != nil {
		return fmt.Errorf("approve-prd: load project: %w", err)
	}

	// auto_start: if project has AutoStart, defer PlanPRDJob until after commit.
	if project.AutoStart {
		jobs.Defer(ctx, &factoryjobs.PlanPRDJob{PRDID: prd.ID})
	}

	return nil
}
