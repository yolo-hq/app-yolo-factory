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

// AdvisorRun triggers an advisor analysis directly from the CLI.
type AdvisorRun struct {
	command.Base
}

type AdvisorRunInput struct {
	Project  string `flag:"project" validate:"required" usage:"Project ID or name"`
	Analysis string `flag:"analysis" usage:"Analysis type: pattern_extraction, code_quality, performance, architecture, model_optimization"`
}

func (c *AdvisorRun) Name() string        { return "advisor:run" }
func (c *AdvisorRun) Description() string { return "Run advisor analysis on a project" }
func (c *AdvisorRun) Input() any          { return &AdvisorRunInput{} }

func (c *AdvisorRun) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*AdvisorRunInput)

	analysis := input.Analysis
	if analysis == "" {
		analysis = "code_quality"
	}

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get project repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Project])

	project, err := helpers.FindProjectByIDOrName(ctx, r, input.Project)
	if err != nil {
		return err
	}

	// Advisor needs Claude client for analysis. Run inline with nil Claude — will fail
	// at agent call but shows the intent. Use worker for full execution.
	// Wire Claude client when command DI is available (tracked in factory issues).
	svc := &services.AdvisorService{}

	cctx.Print("Running advisor %s analysis on %s...", analysis, project.Name)
	out, err := svc.Execute(ctx, services.AdvisorInput{
		ProjectID:    project.ID,
		AnalysisType: analysis,
	})
	if err != nil {
		cctx.Print("ERROR: %s", err)
		cctx.Print("Advisor requires Claude client. Use the worker for full analysis.")
		return nil
	}

	for _, s := range out.Suggestions {
		cctx.Print("  [%s] %s: %s", s.Priority, s.Category, s.Title)
	}

	return nil
}
