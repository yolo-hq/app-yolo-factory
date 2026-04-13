package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers/lint"
)

// Lint runs static analysis checks on a project directory.
type Lint struct {
	command.Base
}

type LintInput struct {
	Path string `flag:"path" validate:"required" usage:"Project root path to lint"`
}

func (c *Lint) Name() string        { return "lint" }
func (c *Lint) Description() string { return "Run static analysis checks on project code" }
func (c *Lint) Input() any          { return &LintInput{} }

func (c *Lint) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*LintInput)

	result, err := lint.RunAll(lint.Options{Path: input.Path})
	if err != nil {
		return fmt.Errorf("lint: %w", err)
	}

	cctx.Print("Lint: %d checks, %d passed, %d failed", result.ChecksRun, result.ChecksPassed, result.ChecksFailed)

	for _, f := range result.Findings {
		cctx.Print("  [%s] %s:%d %s (%s)", f.Severity, f.File, f.Line, f.Message, f.Check)
	}

	if !result.Passed {
		return fmt.Errorf("lint failed with %d error(s)", result.ChecksFailed)
	}

	return nil
}
