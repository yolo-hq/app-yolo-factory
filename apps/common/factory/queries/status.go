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

	SpentThisMonthUSD float64 `field:"spentThisMonthUsd"`
}

// activePRDSummary summarizes a single in-progress PRD.
type activePRDSummary struct {
	Title    string `json:"title"`
	Progress int    `json:"progress"`
}

// StatusResponse is the typed response for StatusQuery.
type StatusResponse struct {
	TasksByStatus   map[string]int     `json:"tasksByStatus"`
	ActivePRDs      []activePRDSummary `json:"activePrds"`
	MonthlySpendUSD float64            `json:"monthlySpendUsd"`
}

// StatusQuery shows a factory dashboard summary.
type StatusQuery struct {
	query.Base
	query.Returns[StatusResponse]
}

func (q *StatusQuery) Description() string { return "Factory dashboard summary" }

func (q *StatusQuery) Execute(ctx context.Context, qctx *query.Context) error {
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
		activePRDs = append(activePRDs, activePRDSummary{Title: p.Title})
	}

	projects, err := read.FindMany[projectSpendRow](ctx)
	if err != nil {
		return fmt.Errorf("status: list projects: %w", err)
	}
	var totalSpent float64
	for _, p := range projects {
		totalSpent += p.SpentThisMonthUSD
	}

	return q.Respond(qctx, StatusResponse{
		TasksByStatus:   byStatus,
		ActivePRDs:      activePRDs,
		MonthlySpendUSD: totalSpent,
	})
}
