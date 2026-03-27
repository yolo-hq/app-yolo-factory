package actions

import (
	"context"

	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type UpdateTaskInput struct {
	Title       *string  `json:"title"`
	Body        *string  `json:"body"`
	Type        *string  `json:"type"`
	Status      *string  `json:"status"`
	Priority    *int     `json:"priority"`
	Model       *string  `json:"model"`
	Labels      *string  `json:"labels"`
	ParentID    *string  `json:"parentId"`
	DependsOn   *string  `json:"dependsOn"`
	MaxRetries  *int     `json:"maxRetries"`
	TimeoutSecs *int     `json:"timeoutSecs"`
}

type UpdateTaskAction struct {
	action.TypedInput[UpdateTaskInput]
	Repo entity.WriteRepository[entities.Task]
}

func (a *UpdateTaskAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}

func (a *UpdateTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.RequireEntityID(actx, "Task"); r != nil {
		return *r
	}
	input, r := a.Input(actx)
	if r != nil {
		return *r
	}

	builder := a.Repo.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: actx.EntityID}).
		SetFromInput(input).
		Returning()

	updated, err := builder.Exec(ctx)
	if err != nil {
		return action.Failure("update failed: " + err.Error())
	}
	if updated == nil {
		return action.NotFound("Task", actx.EntityID)
	}

	return action.Success(updated, "updated")
}
