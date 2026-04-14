package jobs

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/jobs"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// ResetBudgetsJob resets spent_this_month_usd for all active projects.
// Scheduled on the 1st of each month.
type ResetBudgetsJob struct {
	jobs.Base
}

func (j *ResetBudgetsJob) Name() string { return "factory.reset-monthly-budgets" }

func (j *ResetBudgetsJob) Config() jobs.Config {
	return jobs.Config{
		Queue:   "default",
		Timeout: 30 * time.Second,
	}
}

func (j *ResetBudgetsJob) Handle(ctx context.Context, _ []byte) error {
	_, err := svc.S.BudgetReset.Execute(ctx, services.BudgetResetInput{})
	return err
}

func (j *ResetBudgetsJob) Description() string {
	return "Reset monthly budget counters for all projects"
}
