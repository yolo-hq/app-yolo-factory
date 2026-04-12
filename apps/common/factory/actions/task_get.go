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

// TaskDetail is the projection used for single-task responses.
type TaskDetail struct {
	projection.For[entities.Task]

	ID                 string  `field:"id"`
	Title              string  `field:"title"`
	Status             string  `field:"status"`
	Sequence           int     `field:"sequence"`
	Branch             string  `field:"branch"`
	Model              string  `field:"model"`
	CostUSD            float64 `field:"costUsd"`
	RunCount           int     `field:"runCount"`
	MaxRetries         int     `field:"maxRetries"`
	Summary            string  `field:"summary"`
	Spec               string  `field:"spec"`
	AcceptanceCriteria string  `field:"acceptanceCriteria"`
}

// GetTaskResponse is the typed response for GetTaskAction.
type GetTaskResponse struct {
	Task *TaskDetail `json:"task"`
}

// GetTaskAction fetches a single task by ID.
type GetTaskAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.GetTaskInput]
	action.TypedResponse[GetTaskResponse]
}

func (a *GetTaskAction) Description() string { return "Get a task by ID" }

func (a *GetTaskAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	task, err := read.FindOne[TaskDetail](ctx, input.ID)
	if err != nil {
		return fmt.Errorf("get-task: %w", err)
	}
	if task.ID == "" {
		return fmt.Errorf("task %s not found", input.ID)
	}

	return a.Respond(actx, GetTaskResponse{Task: &task})
}
