package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// SentinelRun triggers a sentinel health-check run.
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

func (c *SentinelRun) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*SentinelRunInput)

	if !input.All && input.Project == "" {
		return fmt.Errorf("specify --project or --all")
	}

	if input.All {
		repo, err := cctx.RepoProvider.Repo("Project")
		if err != nil {
			return fmt.Errorf("get project repo: %w", err)
		}
		r := repo.(entity.ReadRepository[entities.Project])

		result, err := r.FindMany(ctx, entity.FindOptions{
			Filters: []entity.FilterCondition{
				{Field: "status", Operator: entity.OpEq, Value: entities.ProjectActive},
			},
		})
		if err != nil {
			return fmt.Errorf("list projects: %w", err)
		}

		for _, p := range result.Data {
			cctx.Print("Sentinel queued for %s (%s)", p.Name, p.ID)
		}
		cctx.Print("Use the worker to process sentinel jobs.")
		return nil
	}

	cctx.Print("Sentinel queued for project %s", input.Project)
	cctx.Print("Use the worker to process sentinel jobs.")
	return nil
}
