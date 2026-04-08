package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

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

	if _, err := w.Update(ctx).WhereID(id).Set(fields.Task.Status.Name(), entities.TaskCancelled).Exec(ctx); err != nil {
		return fmt.Errorf("cancel task: %w", err)
	}

	cctx.Print("Cancelled task %s", id)
	return nil
}
