package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// ResetBudgetsJob resets spent_this_month_usd for all active projects.
// Scheduled on the 1st of each month.
type ResetBudgetsJob struct {
	jobs.Base
	ProjectRead  entity.ReadRepository[entities.Project]
	ProjectWrite entity.WriteRepository[entities.Project]
}

func (j *ResetBudgetsJob) Name() string { return "factory.reset-monthly-budgets" }

func (j *ResetBudgetsJob) Config() jobs.Config {
	return jobs.Config{
		Queue:   "default",
		Timeout: 30 * time.Second,
	}
}

func (j *ResetBudgetsJob) Handle(ctx context.Context, _ []byte) error {
	if j.ProjectRead == nil {
		return fmt.Errorf("factory.reset-monthly-budgets: dependencies not injected")
	}

	result, err := j.ProjectRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: entities.ProjectActive},
		},
	})
	if err != nil {
		return fmt.Errorf("find active projects: %w", err)
	}

	for _, p := range result.Data {
		if _, err := j.ProjectWrite.Update(ctx).
			WhereID(p.ID).
			Set("spent_this_month_usd", 0).
			Exec(ctx); err != nil {
			fmt.Printf("[factory] ERROR: failed to reset budget for project %s: %v\n", p.ID, err)
		}
	}

	return nil
}
