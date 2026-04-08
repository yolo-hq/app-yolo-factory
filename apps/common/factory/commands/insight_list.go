package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type InsightList struct {
	command.Base
}

type InsightListInput struct {
	Project  string `flag:"project" usage:"Filter by project ID"`
	Category string `flag:"category" usage:"Filter by category"`
	Status   string `flag:"status" usage:"Filter by status"`
	Priority string `flag:"priority" usage:"Filter by priority"`
}

func (c *InsightList) Name() string        { return "insight:list" }
func (c *InsightList) Description() string { return "List insights" }
func (c *InsightList) Input() any          { return &InsightListInput{} }

func (c *InsightList) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*InsightListInput)

	repo, err := cctx.RepoProvider.Repo("Insight")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Insight])

	opts := entity.FindOptions{
		Sort: &entity.SortParams{Field: "created_at", Order: "desc"},
	}
	if input != nil {
		if input.Project != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "project_id", Operator: entity.OpEq, Value: input.Project,
			})
		}
		if input.Category != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "category", Operator: entity.OpEq, Value: input.Category,
			})
		}
		if input.Status != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "status", Operator: entity.OpEq, Value: input.Status,
			})
		}
		if input.Priority != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "priority", Operator: entity.OpEq, Value: input.Priority,
			})
		}
	}

	result, err := r.FindMany(ctx, opts)
	if err != nil {
		return fmt.Errorf("list insights: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No insights found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "ID\tTITLE\tCATEGORY\tPRIORITY\tSTATUS\n")
	for _, i := range result.Data {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			i.ID, i.Title, i.Category, i.Priority, i.Status)
	}
	tw.Flush()
	return nil
}
