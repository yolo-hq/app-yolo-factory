package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// TaskStatusCounts groups tasks by status using a single GROUP BY query.
type TaskStatusCounts struct {
	projection.For[entities.Task]
	Status string `field:"status" group:"true"`
	Count  int    `aggregate:"count"`
}

// activePRDRow holds PRD progress data for in-progress PRDs.
type activePRDRow struct {
	projection.For[entities.PRD]

	Title    string `field:"title"`
	Progress int    `field:"progress"`
}

// projectSpendRow holds project spend data.
type projectSpendRow struct {
	projection.For[entities.Project]

	SpentThisMonthUSD float64 `field:"spentThisMonthUsd"`
}

// activePRDSummary summarizes a single in-progress PRD.
type activePRDSummary struct {
	Title    string `json:"title"`
	Progress int    `json:"progress"`
}

// statusResponse is the typed response for StatusAction.
type statusResponse struct {
	TasksByStatus   map[string]int     `json:"tasksByStatus"`
	ActivePRDs      []activePRDSummary `json:"activePrds"`
	MonthlySpendUSD float64            `json:"monthlySpendUsd"`
}

// StatusAction shows a factory dashboard summary.
type StatusAction struct {
	action.TypedResponse[statusResponse]
	action.SkipAllPolicies
}

func (a *StatusAction) ReadOnly() bool      { return true }
func (a *StatusAction) Description() string { return "Factory dashboard summary" }

func (a *StatusAction) Execute(ctx context.Context, actx *action.Context) error {
	counts, err := read.FindMany[TaskStatusCounts](ctx)
	if err != nil {
		return fmt.Errorf("status: task counts: %w", err)
	}
	byStatus := make(map[string]int, len(counts))
	for _, c := range counts {
		byStatus[c.Status] = c.Count
	}

	activePRDRows, err := read.FindMany[activePRDRow](ctx,
		read.Eq("status", string(enums.PRDStatusInProgress)),
	)
	if err != nil {
		return fmt.Errorf("status: list prds: %w", err)
	}
	activePRDs := make([]activePRDSummary, 0, len(activePRDRows))
	for _, p := range activePRDRows {
		activePRDs = append(activePRDs, activePRDSummary{Title: p.Title, Progress: p.Progress})
	}

	projects, err := read.FindMany[projectSpendRow](ctx)
	if err != nil {
		return fmt.Errorf("status: list projects: %w", err)
	}
	var totalSpent float64
	for _, p := range projects {
		totalSpent += p.SpentThisMonthUSD
	}

	return a.Respond(actx, statusResponse{
		TasksByStatus:   byStatus,
		ActivePRDs:      activePRDs,
		MonthlySpendUSD: totalSpent,
	})
}
