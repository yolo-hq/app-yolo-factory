package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type TaskList struct {
	command.Base
}

type TaskListInput struct {
	PRD     string `flag:"prd" usage:"Filter by PRD ID"`
	Project string `flag:"project" usage:"Filter by project ID"`
	Status  string `flag:"status" usage:"Filter by status"`
}

func (c *TaskList) Name() string        { return "task:list" }
func (c *TaskList) Description() string { return "List tasks" }
func (c *TaskList) Input() any          { return &TaskListInput{} }

func (c *TaskList) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*TaskListInput)

	repo, err := cctx.RepoProvider.Repo("Task")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Task])

	opts := entity.FindOptions{
		Sort: &entity.SortParams{Field: "sequence", Order: "asc"},
	}
	if input.PRD != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "prd_id", Operator: entity.OpEq, Value: input.PRD,
		})
	}
	if input.Project != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "project_id", Operator: entity.OpEq, Value: input.Project,
		})
	}
	if input.Status != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "status", Operator: entity.OpEq, Value: input.Status,
		})
	}

	result, err := r.FindMany(ctx, opts)
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No tasks found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "SEQ\tID\tTITLE\tSTATUS\tBRANCH\tCOST\tRUNS\n")
	for _, t := range result.Data {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t$%.2f\t%d\n",
			t.Sequence, t.ID, t.Title, t.Status, t.Branch, t.CostUSD, t.RunCount)
	}
	tw.Flush()
	return nil
}
