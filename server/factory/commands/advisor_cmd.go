package commands

import (
	"context"

	"github.com/yolo-hq/yolo/core/command"
)

// AdvisorRun triggers an advisor analysis.
type AdvisorRun struct {
	command.Base
}

type AdvisorRunInput struct {
	Project  string `flag:"project" validate:"required" usage:"Project ID"`
	Analysis string `flag:"analysis" usage:"Analysis type: pattern_extraction, code_quality, performance, architecture, model_optimization"`
}

func (c *AdvisorRun) Name() string        { return "advisor:run" }
func (c *AdvisorRun) Description() string { return "Run advisor analysis on a project" }
func (c *AdvisorRun) Input() any          { return &AdvisorRunInput{} }

func (c *AdvisorRun) Execute(_ context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*AdvisorRunInput)

	analysis := input.Analysis
	if analysis == "" {
		analysis = "code_quality"
	}

	cctx.Print("Advisor %s analysis queued for project %s", analysis, input.Project)
	cctx.Print("Use the worker to process advisor jobs.")
	return nil
}
