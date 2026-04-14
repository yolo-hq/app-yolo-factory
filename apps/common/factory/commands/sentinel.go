package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/filter"
	"github.com/yolo-hq/yolo/core/read"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// SentinelRunInput is the CLI input for sentinel:run.
type SentinelRunInput struct {
	Project string `flag:"project" usage:"Project ID"`
	All     bool   `flag:"all" usage:"Run against all active projects"`
}

// defaultWatches is the set of watches to run when no specific watch is requested.
var defaultWatches = []string{"build_health", "test_health", "security", "convention_drift"}

// SentinelRun runs sentinel health checks.
//
// @name sentinel:run
func SentinelRun(ctx context.Context, cctx *command.Context, in SentinelRunInput) error {
	if !in.All && in.Project == "" {
		return fmt.Errorf("specify --project or --all")
	}

	var projects []entities.Project

	if in.All {
		result, err := read.FindMany[entities.Project](ctx, filter.Eq(fields.Project.Status.Name(), string(enums.ProjectStatusActive)))
		if err != nil {
			return fmt.Errorf("list projects: %w", err)
		}
		projects = result
	} else {
		p, err := helpers.FindProjectByIDOrName(ctx, in.Project)
		if err != nil {
			return err
		}
		projects = append(projects, *p)
	}

	svc := &services.SentinelService{}

	for _, p := range projects {
		cctx.Print("Running sentinel on %s (%s)...", p.Name, p.ID)
		out, err := svc.Execute(ctx, services.SentinelInput{
			ProjectID: p.ID,
			Watches:   defaultWatches,
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
