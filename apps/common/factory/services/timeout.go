package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// TimeoutService finds running runs that have exceeded their timeout and fails them.
type TimeoutService struct {
	service.Base
	RunRead     entity.ReadRepository[entities.Run]
	RunWrite    entity.WriteRepository[entities.Run]
	TaskRead    entity.ReadRepository[entities.Task]
	TaskWrite   entity.WriteRepository[entities.Task]
	ProjectRead entity.ReadRepository[entities.Project]
}

// TimeoutInput is empty — this service scans all running runs.
type TimeoutInput struct{}

// TimeoutOutput reports how many runs were timed out.
type TimeoutOutput struct {
	TimedOut int
}

// Execute checks all running runs for timeout and fails them, retrying tasks if under limit.
func (s *TimeoutService) Execute(ctx context.Context, _ TimeoutInput) (TimeoutOutput, error) {
	var out TimeoutOutput

	// Find all running runs.
	result, err := s.RunRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: fields.Run.Status.Name(), Operator: entity.OpEq, Value: string(enums.RunStatusRunning)},
		},
	})
	if err != nil {
		return out, fmt.Errorf("find running runs: %w", err)
	}

	for _, run := range result.Data {
		// Load the task to get timeout and retry config.
		task, err := s.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: run.TaskID})
		if err != nil || task == nil {
			continue
		}

		// Load project for timeout setting.
		project, err := s.ProjectRead.FindOne(ctx, entity.FindOneOptions{ID: task.ProjectID})
		if err != nil || project == nil {
			continue
		}

		timeoutSecs := project.TimeoutSecs
		if timeoutSecs <= 0 {
			timeoutSecs = 600 // default 10 minutes
		}

		if time.Since(run.StartedAt) <= time.Duration(timeoutSecs)*time.Second {
			continue
		}

		// Timed out — fail the run.
		now := time.Now()
		if _, err := s.RunWrite.Update(ctx).
			WhereID(run.ID).
			Set(fields.Run.Status.Name(), string(enums.RunStatusFailed)).
			Set(fields.Run.Error.Name(), "timeout").
			Set(fields.Run.CompletedAt.Name(), now).
			Exec(ctx); err != nil {
			slog.Error("failed to timeout run", "run_id", run.ID, "error", err)
			continue
		}

		out.TimedOut++

		// Update task: retry if under limit, else fail.
		newRunCount := task.RunCount + 1
		if newRunCount < task.MaxRetries {
			if _, err := s.TaskWrite.Update(ctx).
				WhereID(task.ID).
				Set(fields.Task.Status.Name(), string(enums.TaskStatusQueued)).
				Set(fields.Task.RunCount.Name(), newRunCount).
				Exec(ctx); err != nil {
				slog.Error("failed to requeue task", "task_id", task.ID, "error", err)
			}
		} else {
			if _, err := s.TaskWrite.Update(ctx).
				WhereID(task.ID).
				Set(fields.Task.Status.Name(), string(enums.TaskStatusFailed)).
				Set(fields.Task.RunCount.Name(), newRunCount).
				Exec(ctx); err != nil {
				slog.Error("failed to fail task", "task_id", task.ID, "error", err)
			}
		}
	}

	return out, nil
}

func (s *TimeoutService) Description() string { return "Check and fail timed-out runs" }
