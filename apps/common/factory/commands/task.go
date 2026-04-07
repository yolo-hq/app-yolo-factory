package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// --- TaskList ---

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

// --- TaskGet ---

type TaskGet struct {
	command.Base
}

func (c *TaskGet) Name() string        { return "task:get" }
func (c *TaskGet) Description() string { return "Get a task by ID" }

func (c *TaskGet) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: task:get <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Task")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Task])

	t, err := r.FindOne(ctx, entity.FindOneOptions{ID: id})
	if err != nil {
		return fmt.Errorf("find task: %w", err)
	}
	if t == nil {
		return fmt.Errorf("task %s not found", id)
	}

	cctx.Print("ID:       %s", t.ID)
	cctx.Print("Title:    %s", t.Title)
	cctx.Print("Status:   %s", t.Status)
	cctx.Print("Sequence: %d", t.Sequence)
	cctx.Print("Branch:   %s", t.Branch)
	cctx.Print("Model:    %s", t.Model)
	cctx.Print("Cost:     $%.2f", t.CostUSD)
	cctx.Print("Runs:     %d/%d", t.RunCount, t.MaxRetries)
	if t.Summary != "" {
		cctx.Print("Summary:  %s", t.Summary)
	}
	return nil
}

// --- TaskCancel ---

type TaskCancel struct {
	command.Base
}

func (c *TaskCancel) Name() string        { return "task:cancel" }
func (c *TaskCancel) Description() string { return "Cancel a task" }

func (c *TaskCancel) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: task:cancel <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Task")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Task])

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.TaskCancelled).Exec(ctx); err != nil {
		return fmt.Errorf("cancel task: %w", err)
	}

	cctx.Print("Cancelled task %s", id)
	return nil
}

// --- TaskRetry ---

type TaskRetry struct {
	command.Base
}

type TaskRetryInput struct {
	Model string `flag:"model" usage:"Override model for retry"`
}

func (c *TaskRetry) Name() string        { return "task:retry" }
func (c *TaskRetry) Description() string { return "Retry a failed task" }
func (c *TaskRetry) Input() any          { return &TaskRetryInput{} }

func (c *TaskRetry) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: task:retry <id> [--model]")
	}
	id := cctx.Args[0]
	input, _ := cctx.TypedInput.(*TaskRetryInput)

	repo, err := cctx.RepoProvider.Repo("Task")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Task])

	ub := w.Update(ctx).WhereID(id).Set("status", entities.TaskQueued)
	if input != nil && input.Model != "" {
		ub = ub.Set("model", input.Model)
	}

	if _, err := ub.Exec(ctx); err != nil {
		return fmt.Errorf("retry task: %w", err)
	}

	cctx.Print("Retrying task %s", id)
	return nil
}
