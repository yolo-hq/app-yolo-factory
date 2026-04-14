package commands

import (
	"context"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// AdvisorRunInput is the CLI input for advisor:run.
type AdvisorRunInput struct {
	Project  string `flag:"project" validate:"required" usage:"Project ID or name"`
	Analysis string `flag:"analysis" usage:"Analysis type: pattern_extraction, code_quality, performance, architecture, model_optimization"`
}

// AdvisorRun runs advisor analysis on a project.
//
// @name advisor:run
func AdvisorRun(ctx context.Context, cctx *command.Context, in AdvisorRunInput) error {
	analysis := in.Analysis
	if analysis == "" {
		analysis = "code_quality"
	}

	project, err := helpers.FindProjectByIDOrName(ctx, in.Project)
	if err != nil {
		return err
	}

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
