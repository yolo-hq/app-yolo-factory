package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

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

	ub := w.Update(ctx).WhereID(id).Set(fields.Task.Status.Name(), string(enums.TaskStatusQueued))
	if input != nil && input.Model != "" {
		ub = ub.Set(fields.Task.Model.Name(), input.Model)
	}

	if _, err := ub.Exec(ctx); err != nil {
		return fmt.Errorf("retry task: %w", err)
	}

	cctx.Print("Retrying task %s", id)
	return nil
}
