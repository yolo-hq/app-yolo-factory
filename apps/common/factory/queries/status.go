package queries

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/query"
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

	Title string `field:"title"`
}

// projectSpendRow holds project spend data.
type projectSpendRow struct {
	projection.For[entities.Project]

	SpentThisMonthUSD float64 `field:"spent_this_month_usd"`
}

// activePRDSummary summarizes a single in-progress PRD.
type activePRDSummary struct {
	Title    string `json:"title"`
	Progress int    `json:"progress"`
}

// StatusInput is an empty input marker for the factory status query.
type StatusInput struct{}

// StatusResponse is the typed response for StatusQuery.
type StatusResponse struct {
	TasksByStatus   map[string]int     `json:"tasksByStatus"`
	ActivePRDs      []activePRDSummary `json:"activePrds"`
	MonthlySpendUSD float64            `json:"monthlySpendUsd"`
}

// Status shows a factory dashboard summary.
func Status(ctx context.Context, qctx *query.Context, in StatusInput) (StatusResponse, error) {
	_ = in

	counts, err := read.FindMany[TaskStatusCounts](ctx)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("status: task counts: %w", err)
	}
	byStatus := make(map[string]int, len(counts))
	for _, c := range counts {
		byStatus[c.Status] = c.Count
	}

	activePRDRows, err := read.FindMany[activePRDRow](ctx,
		read.Eq("status", string(enums.PRDStatusInProgress)),
	)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("status: list prds: %w", err)
	}
	activePRDs := make([]activePRDSummary, 0, len(activePRDRows))
	for _, p := range activePRDRows {
		activePRDs = append(activePRDs, activePRDSummary{Title: p.Title})
	}

	projects, err := read.FindMany[projectSpendRow](ctx)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("status: list projects: %w", err)
	}
	var totalSpent float64
	for _, p := range projects {
		totalSpent += p.SpentThisMonthUSD
	}

	return StatusResponse{
		TasksByStatus:   byStatus,
		ActivePRDs:      activePRDs,
		MonthlySpendUSD: totalSpent,
	}, nil
}
