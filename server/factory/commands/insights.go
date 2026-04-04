package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// --- InsightList ---

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

// --- InsightAcknowledge ---

type InsightAcknowledge struct {
	command.Base
}

func (c *InsightAcknowledge) Name() string        { return "insight:acknowledge" }
func (c *InsightAcknowledge) Description() string { return "Acknowledge an insight" }
func (c *InsightAcknowledge) Input() any          { return nil }

func (c *InsightAcknowledge) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: insight:acknowledge <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Insight")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Insight])

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.InsightAcknowledged).Exec(ctx); err != nil {
		return fmt.Errorf("acknowledge insight: %w", err)
	}

	cctx.Print("Acknowledged insight %s", id)
	return nil
}

// --- InsightApply ---

type InsightApply struct {
	command.Base
}

func (c *InsightApply) Name() string        { return "insight:apply" }
func (c *InsightApply) Description() string { return "Mark an insight as applied" }
func (c *InsightApply) Input() any          { return nil }

func (c *InsightApply) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: insight:apply <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Insight")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Insight])

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.InsightApplied).Exec(ctx); err != nil {
		return fmt.Errorf("apply insight: %w", err)
	}

	cctx.Print("Applied insight %s", id)
	return nil
}

// --- InsightDismiss ---

type InsightDismiss struct {
	command.Base
}

type InsightDismissInput struct {
	Reason string `flag:"reason" validate:"required" usage:"Reason for dismissal"`
}

func (c *InsightDismiss) Name() string        { return "insight:dismiss" }
func (c *InsightDismiss) Description() string { return "Dismiss an insight" }
func (c *InsightDismiss) Input() any          { return &InsightDismissInput{} }

func (c *InsightDismiss) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: insight:dismiss <id> --reason <reason>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Insight")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Insight])

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.InsightDismissed).Exec(ctx); err != nil {
		return fmt.Errorf("dismiss insight: %w", err)
	}

	cctx.Print("Dismissed insight %s", id)
	return nil
}
