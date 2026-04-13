package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// BudgetResetService resets monthly budget counters for all active projects.
type BudgetResetService struct {
	service.Base
	ProjectWrite entity.WriteRepository[entities.Project]
}

// BudgetResetInput is empty — this service resets all active projects.
type BudgetResetInput struct{}

// BudgetResetOutput reports how many projects were reset.
type BudgetResetOutput struct {
	ResetCount int
}

// Execute resets spent_this_month_usd to 0 for all active projects.
func (s *BudgetResetService) Execute(ctx context.Context, _ BudgetResetInput) (BudgetResetOutput, error) {
	var out BudgetResetOutput

	projects, err := read.FindMany[entities.Project](ctx,
		read.Eq(fields.Project.Status.Name(), string(enums.ProjectStatusActive)),
		read.Limit(1000),
	)
	if err != nil {
		return out, fmt.Errorf("find active projects: %w", err)
	}

	for _, p := range projects {
		if _, err := s.ProjectWrite.Update(ctx).
			WhereID(p.ID).
			Set(fields.Project.SpentThisMonthUSD.Name(), 0).
			Exec(ctx); err != nil {
			slog.Error("failed to reset budget for project", "project_id", p.ID, "error", err)
			continue
		}
		out.ResetCount++
	}

	return out, nil
}

func (s *BudgetResetService) Description() string { return "Reset monthly budget counters" }
