package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type PRDExecute struct {
	command.Base
}

func (c *PRDExecute) Name() string        { return "prd:execute" }
func (c *PRDExecute) Description() string { return "Execute a PRD (triggers planning job)" }

func (c *PRDExecute) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: prd:execute <id>")
	}
	id := cctx.Args[0]

	// Verify PRD exists and is in a valid state.
	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.PRD])
	w := repo.(entity.WriteRepository[entities.PRD])

	prd, err := r.FindOne(ctx, entity.FindOneOptions{ID: id})
	if err != nil {
		return fmt.Errorf("find prd: %w", err)
	}
	if prd == nil {
		return fmt.Errorf("PRD %s not found", id)
	}

	if prd.Status != entities.PRDApproved && prd.Status != entities.PRDDraft {
		return fmt.Errorf("PRD must be in draft or approved status, got %s", prd.Status)
	}

	// Mark PRD as planning to trigger the PlanPRDJob via the worker.
	if _, err := w.Update(ctx).WhereID(id).Set(fields.PRD.Status.Name(), entities.PRDPlanning).Exec(ctx); err != nil {
		return fmt.Errorf("update prd: %w", err)
	}

	cctx.Print("PRD %s marked for planning. Worker will pick up the PlanPRDJob.", id)
	return nil
}
