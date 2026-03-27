package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type CheckTimeoutsJob struct {
	jobs.Base
	RunRead   entity.ReadRepository[entities.Run]
	RunWrite  entity.WriteRepository[entities.Run]
	TaskRead  entity.ReadRepository[entities.Task]
	TaskWrite entity.WriteRepository[entities.Task]
}

func (j *CheckTimeoutsJob) Name() string { return "factory.check-timeouts" }

func (j *CheckTimeoutsJob) Config() jobs.Config {
	return jobs.Config{
		Queue:   "maintenance",
		Timeout: 30 * time.Second,
	}
}

func (j *CheckTimeoutsJob) Handle(ctx context.Context, _ []byte) error {
	running, err := j.RunRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: "running"},
		},
	})
	if err != nil {
		return err
	}

	now := time.Now()
	for _, run := range running.Data {
		task, _ := j.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: run.TaskID})
		if task == nil {
			continue
		}

		timeout := time.Duration(task.TimeoutSecs) * time.Second
		if now.Sub(run.StartedAt) <= timeout {
			continue
		}

		// Mark run as failed
		j.RunWrite.Update(ctx).
			Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: run.ID}).
			Set("status", "failed").
			Set("error", fmt.Sprintf("timeout after %ds", task.TimeoutSecs)).
			Set("completed_at", &now).
			Exec(ctx)

		// Re-queue or fail task
		if task.RunCount < task.MaxRetries {
			j.TaskWrite.Update(ctx).
				Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: task.ID}).
				Set("status", "queued").
				Exec(ctx)
		} else {
			j.TaskWrite.Update(ctx).
				Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: task.ID}).
				Set("status", "failed").
				Exec(ctx)
		}
	}
	return nil
}
