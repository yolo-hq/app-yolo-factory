package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// SentinelRun triggers a sentinel health-check run directly from the CLI.
type SentinelRun struct {
	command.Base
}

type SentinelRunInput struct {
	Project string `flag:"project" usage:"Project ID"`
	All     bool   `flag:"all" usage:"Run against all active projects"`
}

func (c *SentinelRun) Name() string        { return "sentinel:run" }
func (c *SentinelRun) Description() string { return "Run sentinel health checks" }
func (c *SentinelRun) Input() any          { return &SentinelRunInput{} }

// defaultWatches is the set of watches to run when no specific watch is requested.
var defaultWatches = []string{"build_health", "test_health", "security", "convention_drift"}

func (c *SentinelRun) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*SentinelRunInput)

	if !input.All && input.Project == "" {
		return fmt.Errorf("specify --project or --all")
	}

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get project repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Project])

	var projects []entities.Project

	if input.All {
		result, err := r.FindMany(ctx, entity.FindOptions{
			Filters: []entity.FilterCondition{
				{Field: "status", Operator: entity.OpEq, Value: entities.ProjectActive},
			},
		})
		if err != nil {
			return fmt.Errorf("list projects: %w", err)
		}
		projects = result.Data
	} else {
		p, err := helpers.FindProjectByIDOrName(ctx, r, input.Project)
		if err != nil {
			return err
		}
		projects = append(projects, *p)
	}

	// Run sentinel directly (no Claude client needed for build/test/security watches).
	svc := &services.SentinelService{}

	for _, p := range projects {
		cctx.Print("Running sentinel on %s (%s)...", p.Name, p.ID)
		out, err := svc.Execute(ctx, services.SentinelInput{
			Project: p,
			Watches: defaultWatches,
		})
		if err != nil {
			cctx.Print("  ERROR: %s", err)
			continue
		}
		for _, f := range out.Findings {
			cctx.Print("  [%s] %s: %s", f.Severity, f.Watch, f.Message)
		}
		if len(out.TasksToCreate) > 0 {
			cctx.Print("  %d tasks would be created (use worker for persistence)", len(out.TasksToCreate))
		}
		if len(out.SuggestionsToCreate) > 0 {
			cctx.Print("  %d suggestions would be created (use worker for persistence)", len(out.SuggestionsToCreate))
		}
	}

	return nil
}
