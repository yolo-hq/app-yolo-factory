package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers/lint"
)

// LintInput is the CLI input for the lint command.
type LintInput struct {
	Path string `flag:"path" validate:"required" usage:"Project root path to lint"`
}

// Lint runs static analysis checks on project code.
//
// @name lint
func Lint(ctx context.Context, cctx *command.Context, in LintInput) error {
	_ = ctx
	result, err := lint.RunAll(lint.Options{Path: in.Path})
	if err != nil {
		return err
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
