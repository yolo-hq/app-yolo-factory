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

// TaskStatusRow holds task status data.
type TaskStatusRow struct {
	projection.For[entities.Task]

	Status string `field:"status"`
}

// ActivePRDRow holds PRD progress data.
type ActivePRDRow struct {
	projection.For[entities.PRD]

	Title          string `field:"title"`
	CompletedTasks int    `field:"completedTasks"`
	TotalTasks     int    `field:"totalTasks"`
	Status         string `field:"status"`
}

// ProjectSpendRow holds project spend data.
type ProjectSpendRow struct {
	projection.For[entities.Project]

	SpentThisMonthUSD float64 `field:"spentThisMonthUsd"`
}

// TaskCounts summarizes task counts by status.
type TaskCounts struct {
	Queued    int `json:"queued"`
	Running   int `json:"running"`
	Done      int `json:"done"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
	Blocked   int `json:"blocked"`
	Total     int `json:"total"`
}

// ActivePRDSummary summarizes a single in-progress PRD.
type ActivePRDSummary struct {
	Title          string `json:"title"`
	CompletedTasks int    `json:"completedTasks"`
	TotalTasks     int    `json:"totalTasks"`
	PercentDone    int    `json:"percentDone"`
}

// StatusResponse is the typed response for StatusAction.
type StatusResponse struct {
	Tasks          TaskCounts         `json:"tasks"`
	ActivePRDs     []ActivePRDSummary `json:"activePrds"`
	MonthlySpendUSD float64           `json:"monthlySpendUsd"`
}

// StatusAction shows a factory dashboard summary.
type StatusAction struct {
	action.SkipAllPolicies
	action.NoInput
	action.TypedResponse[StatusResponse]
}

func (a *StatusAction) Description() string { return "Show factory status summary" }

func (a *StatusAction) Execute(ctx context.Context, actx *action.Context) error {
	tasks, err := read.FindMany[TaskStatusRow](ctx)
	if err != nil {
		return fmt.Errorf("status: list tasks: %w", err)
	}

	counts := TaskCounts{Total: len(tasks)}
	for _, t := range tasks {
		switch t.Status {
		case string(enums.TaskStatusQueued):
			counts.Queued++
		case string(enums.TaskStatusRunning):
			counts.Running++
		case string(enums.TaskStatusDone):
			counts.Done++
		case string(enums.TaskStatusFailed):
			counts.Failed++
		case string(enums.TaskStatusCancelled):
			counts.Cancelled++
		case string(enums.TaskStatusBlocked):
			counts.Blocked++
		}
	}

	activePRDRows, err := read.FindMany[ActivePRDRow](ctx,
		read.Eq("status", string(enums.PRDStatusInProgress)),
	)
	if err != nil {
		return fmt.Errorf("status: list prds: %w", err)
	}

	activePRDs := make([]ActivePRDSummary, 0, len(activePRDRows))
	for _, p := range activePRDRows {
		pct := 0
		if p.TotalTasks > 0 {
			pct = p.CompletedTasks * 100 / p.TotalTasks
		}
		activePRDs = append(activePRDs, ActivePRDSummary{
			Title:          p.Title,
			CompletedTasks: p.CompletedTasks,
			TotalTasks:     p.TotalTasks,
			PercentDone:    pct,
		})
	}

	projects, err := read.FindMany[ProjectSpendRow](ctx)
	if err != nil {
		return fmt.Errorf("status: list projects: %w", err)
	}

	var totalSpent float64
	for _, p := range projects {
		totalSpent += p.SpentThisMonthUSD
	}

	return a.Respond(actx, StatusResponse{
		Tasks:           counts,
		ActivePRDs:      activePRDs,
		MonthlySpendUSD: totalSpent,
	})
}
