package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// TaskRow is the projection used for list responses.
type TaskRow struct {
	projection.For[entities.Task]

	ID       string  `field:"id"`
	Title    string  `field:"title"`
	Status   string  `field:"status"`
	Sequence int     `field:"sequence"`
	Branch   string  `field:"branch"`
	CostUSD  float64 `field:"costUsd"`
	RunCount int     `field:"runCount"`
}

// ListTasksResponse is the typed response for ListTasksAction.
type ListTasksResponse struct {
	Tasks []TaskRow `json:"tasks"`
}

// ListTasksAction lists tasks with optional filters.
type ListTasksAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.ListTasksInput]
	action.TypedResponse[ListTasksResponse]
}

func (a *ListTasksAction) Description() string { return "List tasks with optional filters" }

func (a *ListTasksAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{read.OrderBy("sequence", read.Asc)}
	if input.PRDID != "" {
		opts = append(opts, read.Eq("prd_id", input.PRDID))
	}
	if input.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", input.ProjectID))
	}
	if input.Status != "" {
		opts = append(opts, read.Eq("status", input.Status))
	}

	tasks, err := read.FindMany[TaskRow](ctx, opts...)
	if err != nil {
		return fmt.Errorf("list-tasks: %w", err)
	}

	return a.Respond(actx, ListTasksResponse{Tasks: tasks})
}
