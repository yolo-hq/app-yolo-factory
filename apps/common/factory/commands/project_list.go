package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type ProjectList struct {
	command.Base
}

func (c *ProjectList) Name() string        { return "project:list" }
func (c *ProjectList) Description() string { return "List all projects" }

func (c *ProjectList) Execute(ctx context.Context, cctx command.Context) error {
	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Project])

	result, err := r.FindMany(ctx, entity.FindOptions{
		Sort: &entity.SortParams{Field: "name", Order: "asc"},
	})
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No projects found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "ID\tNAME\tSTATUS\tMODEL\tBUDGET\tSPENT\n")
	for _, p := range result.Data {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t$%.2f\t$%.2f\n",
			p.ID, p.Name, p.Status, p.DefaultModel, p.BudgetMonthlyUSD, p.SpentThisMonthUSD)
	}
	tw.Flush()
	return nil
}
