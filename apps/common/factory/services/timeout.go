package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// TimeoutService finds running runs that have exceeded their timeout and fails them.
type TimeoutService struct {
	service.Base
	RunWrite  entity.WriteRepository[entities.Run]
	TaskWrite entity.WriteRepository[entities.Task]
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
	runs, err := read.FindMany[entities.Run](ctx,
		read.Eq(fields.Run.Status.Name(), string(enums.RunStatusRunning)),
		read.Limit(1000),
	)
	if err != nil {
		return out, fmt.Errorf("find running runs: %w", err)
	}

	for _, run := range runs {
		// Load the task to get timeout and retry config.
		task, err := read.FindOne[entities.Task](ctx, run.TaskID)
		if err != nil || task.ID == "" {
			continue
		}

		// Load project for timeout setting.
		project, err := read.FindOne[entities.Project](ctx, task.ProjectID)
		if err != nil || project.ID == "" {
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
