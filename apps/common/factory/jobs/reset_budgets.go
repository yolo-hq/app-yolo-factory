package jobs

import (
	"context"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// ResetBudgetsPayload is an empty payload marker.
type ResetBudgetsPayload struct{}

// ResetBudgets resets monthly budget counters for all projects.
//
// @name factory.reset-monthly-budgets
// @queue default
// @timeout 30s
func ResetBudgets(ctx context.Context, _ ResetBudgetsPayload) error {
	_, err := svc.S.BudgetReset.Execute(ctx, services.BudgetResetInput{})
	return err
}
