package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type ProjectGet struct {
	command.Base
}

func (c *ProjectGet) Name() string        { return "project:get" }
func (c *ProjectGet) Description() string { return "Get a project by ID or name" }

func (c *ProjectGet) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:get <id-or-name>")
	}
	idOrName := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Project])

	p, err := findProjectByIDOrName(ctx, r, idOrName)
	if err != nil {
		return err
	}

	cctx.Print("ID:       %s", p.ID)
	cctx.Print("Name:     %s", p.Name)
	cctx.Print("Status:   %s", p.Status)
	cctx.Print("Repo:     %s", p.RepoURL)
	cctx.Print("Path:     %s", p.LocalPath)
	cctx.Print("Branch:   %s", p.DefaultBranch)
	cctx.Print("Model:    %s", p.DefaultModel)
	cctx.Print("Budget:   $%.2f/month", p.BudgetMonthlyUSD)
	cctx.Print("Spent:    $%.2f", p.SpentThisMonthUSD)
	return nil
}
