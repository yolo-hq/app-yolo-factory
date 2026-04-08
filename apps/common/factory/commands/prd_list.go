package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type PRDList struct {
	command.Base
}

type PRDListInput struct {
	Project string `flag:"project" usage:"Filter by project ID or name"`
	Status  string `flag:"status" usage:"Filter by status"`
}

func (c *PRDList) Name() string        { return "prd:list" }
func (c *PRDList) Description() string { return "List PRDs" }
func (c *PRDList) Input() any          { return &PRDListInput{} }

func (c *PRDList) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*PRDListInput)

	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.PRD])

	opts := entity.FindOptions{
		Sort: &entity.SortParams{Field: "created_at", Order: "desc"},
	}
	if input.Status != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "status", Operator: entity.OpEq, Value: input.Status,
		})
	}
	if input.Project != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "project_id", Operator: entity.OpEq, Value: input.Project,
		})
	}

	result, err := r.FindMany(ctx, opts)
	if err != nil {
		return fmt.Errorf("list prds: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No PRDs found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "ID\tTITLE\tSTATUS\tTASKS\tCOST\n")
	for _, p := range result.Data {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d/%d\t$%.2f\n",
			p.ID, p.Title, p.Status, p.CompletedTasks, p.TotalTasks, p.TotalCostUSD)
	}
	tw.Flush()
	return nil
}
