package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

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
