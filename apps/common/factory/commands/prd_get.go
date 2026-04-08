package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type PRDGet struct {
	command.Base
}

func (c *PRDGet) Name() string        { return "prd:get" }
func (c *PRDGet) Description() string { return "Get a PRD by ID" }

func (c *PRDGet) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: prd:get <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.PRD])

	p, err := r.FindOne(ctx, entity.FindOneOptions{ID: id})
	if err != nil {
		return fmt.Errorf("find prd: %w", err)
	}
	if p == nil {
		return fmt.Errorf("PRD %s not found", id)
	}

	cctx.Print("ID:       %s", p.ID)
	cctx.Print("Title:    %s", p.Title)
	cctx.Print("Status:   %s", p.Status)
	cctx.Print("Source:   %s", p.Source)
	cctx.Print("Tasks:    %d/%d completed", p.CompletedTasks, p.TotalTasks)
	cctx.Print("Cost:     $%.2f", p.TotalCostUSD)
	cctx.Print("")
	cctx.Print("Body:")
	cctx.Print("%s", p.Body)
	return nil
}
