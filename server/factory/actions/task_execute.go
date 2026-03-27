package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// ExecuteTaskAction picks the highest-priority queued auto task and starts execution.
type ExecuteTaskAction struct {
	action.NoInput
	TaskRead  entity.ReadRepository[entities.Task]
	TaskWrite entity.WriteRepository[entities.Task]
	RunWrite  entity.WriteRepository[entities.Run]
	RepoRead  entity.ReadRepository[entities.Repo]
}

func (a *ExecuteTaskAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}

func (a *ExecuteTaskAction) StatusCode() int { return 202 }

func (a *ExecuteTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	// Find highest priority queued auto task
	tasks, err := a.TaskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: "queued"},
			{Field: "type", Operator: entity.OpEq, Value: "auto"},
		},
		Sort:       &entity.SortParams{Field: "priority", Order: "desc"},
		Pagination: &entity.PaginationParams{Limit: 1},
	})
	if err != nil {
		return action.InternalError()
	}
	if len(tasks.Data) == 0 {
		return action.Failure("no queued auto tasks")
	}

	task := tasks.Data[0]

	// Load repo
	repo, err := a.RepoRead.FindOne(ctx, entity.FindOneOptions{ID: task.RepoID})
	if err != nil || repo == nil {
		return action.Failure(fmt.Sprintf("repo %s not found", task.RepoID))
	}
	if !repo.Active {
		return action.Failure(fmt.Sprintf("repo %s is inactive", repo.Name))
	}

	// Determine model
	model := task.Model
	if model == "" {
		model = repo.DefaultModel
	}
	if model == "" {
		model = "sonnet"
	}

	// Update task status
	a.TaskWrite.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: task.ID}).
		Set("status", "running").
		Set("run_count", task.RunCount+1).
		Exec(ctx)

	// Create run
	now := time.Now()
	run := &entities.Run{
		TaskID:    task.ID,
		RepoID:    task.RepoID,
		Agent:     "claude-cli",
		Model:     model,
		Status:    "running",
		StartedAt: now,
	}
	run.ID = ulid.Make().String()

	created, err := a.RunWrite.Insert(ctx, run)
	if err != nil {
		return action.InternalError()
	}

	return action.Success(map[string]any{
		"taskId": task.ID,
		"runId":  created.ID,
		"repo":   repo.Name,
		"model":  model,
	}, "task execution started")
}
