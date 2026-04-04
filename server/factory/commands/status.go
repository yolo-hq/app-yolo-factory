package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// Status shows a summary of active tasks, PRDs, and costs.
type Status struct {
	command.Base
}

func (c *Status) Name() string        { return "status" }
func (c *Status) Description() string { return "Show factory status summary" }

func (c *Status) Execute(ctx context.Context, cctx command.Context) error {
	// Task counts by status.
	taskRepo, err := cctx.RepoProvider.Repo("Task")
	if err != nil {
		return fmt.Errorf("get task repo: %w", err)
	}
	tr := taskRepo.(entity.ReadRepository[entities.Task])

	allTasks, err := tr.FindMany(ctx, entity.FindOptions{})
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	counts := map[string]int{}
	for _, t := range allTasks.Data {
		counts[t.Status]++
	}

	cctx.Print("=== Factory Status ===")
	cctx.Print("")
	cctx.Print("Tasks:")
	for _, s := range []string{entities.TaskQueued, entities.TaskRunning, entities.TaskDone, entities.TaskFailed, entities.TaskCancelled, entities.TaskBlocked} {
		if c := counts[s]; c > 0 {
			cctx.Print("  %-12s %d", s, c)
		}
	}
	cctx.Print("  %-12s %d", "total", len(allTasks.Data))

	// Active PRDs with progress.
	prdRepo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get prd repo: %w", err)
	}
	pr := prdRepo.(entity.ReadRepository[entities.PRD])

	activePRDs, err := pr.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: entities.PRDInProgress},
		},
	})
	if err != nil {
		return fmt.Errorf("list active prds: %w", err)
	}

	if len(activePRDs.Data) > 0 {
		cctx.Print("")
		cctx.Print("Active PRDs:")
		for _, p := range activePRDs.Data {
			pct := 0
			if p.TotalTasks > 0 {
				pct = p.CompletedTasks * 100 / p.TotalTasks
			}
			cctx.Print("  %s — %d/%d tasks (%d%%)", p.Title, p.CompletedTasks, p.TotalTasks, pct)
		}
	}

	// Cost summary from projects.
	projectRepo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get project repo: %w", err)
	}
	projR := projectRepo.(entity.ReadRepository[entities.Project])

	projects, err := projR.FindMany(ctx, entity.FindOptions{})
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	var totalSpent float64
	for _, p := range projects.Data {
		totalSpent += p.SpentThisMonthUSD
	}

	cctx.Print("")
	cctx.Print("Cost this month: $%.2f", totalSpent)

	return nil
}
