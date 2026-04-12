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

// taskStatusRow holds task status data.
type taskStatusRow struct {
	projection.For[entities.Task]

	Status string `field:"status"`
}

// activePRDRow holds PRD progress data.
type activePRDRow struct {
	projection.For[entities.PRD]

	Title          string `field:"title"`
	CompletedTasks int    `field:"completedTasks"`
	TotalTasks     int    `field:"totalTasks"`
	Status         string `field:"status"`
}

// projectSpendRow holds project spend data.
type projectSpendRow struct {
	projection.For[entities.Project]

	SpentThisMonthUSD float64 `field:"spentThisMonthUsd"`
}

// taskCounts summarizes task counts by status.
type taskCounts struct {
	Queued    int `json:"queued"`
	Running   int `json:"running"`
	Done      int `json:"done"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
	Blocked   int `json:"blocked"`
	Total     int `json:"total"`
}

// activePRDSummary summarizes a single in-progress PRD.
type activePRDSummary struct {
	Title          string `json:"title"`
	CompletedTasks int    `json:"completedTasks"`
	TotalTasks     int    `json:"totalTasks"`
	PercentDone    int    `json:"percentDone"`
}

// statusResponse is the typed response for StatusAction.
type statusResponse struct {
	Tasks           taskCounts         `json:"tasks"`
	ActivePRDs      []activePRDSummary `json:"activePrds"`
	MonthlySpendUSD float64            `json:"monthlySpendUsd"`
}

// StatusAction shows a factory dashboard summary.
type StatusAction struct {
	action.TypedResponse[statusResponse]
	action.SkipAllPolicies
}

func (a *StatusAction) ReadOnly() bool     { return true }
func (a *StatusAction) Description() string { return "Factory dashboard summary" }

func (a *StatusAction) Execute(ctx context.Context, actx *action.Context) error {
	tasks, err := read.FindMany[taskStatusRow](ctx)
	if err != nil {
		return fmt.Errorf("status: list tasks: %w", err)
	}

	counts := taskCounts{Total: len(tasks)}
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

	activePRDRows, err := read.FindMany[activePRDRow](ctx,
		read.Eq("status", string(enums.PRDStatusInProgress)),
	)
	if err != nil {
		return fmt.Errorf("status: list prds: %w", err)
	}

	activePRDs := make([]activePRDSummary, 0, len(activePRDRows))
	for _, p := range activePRDRows {
		pct := 0
		if p.TotalTasks > 0 {
			pct = p.CompletedTasks * 100 / p.TotalTasks
		}
		activePRDs = append(activePRDs, activePRDSummary{
			Title:          p.Title,
			CompletedTasks: p.CompletedTasks,
			TotalTasks:     p.TotalTasks,
			PercentDone:    pct,
		})
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
		Tasks:           counts,
		ActivePRDs:      activePRDs,
		MonthlySpendUSD: totalSpent,
	})
}
