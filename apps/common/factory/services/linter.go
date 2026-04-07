package services

import (
	"context"

	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/lint"
)

// LinterService runs static analysis checks on project code.
type LinterService struct {
	service.Base
}

// LinterInput configures a lint run.
type LinterInput struct {
	ProjectPath  string
	ChangedFiles []string
}

// LinterOutput holds lint results.
type LinterOutput struct {
	Passed       bool
	ChecksRun    int
	ChecksPassed int
	ChecksFailed int
	Findings     []lint.Finding
}

// Execute runs all lint checks and returns results.
func (s *LinterService) Execute(_ context.Context, input LinterInput) (LinterOutput, error) {
	result, err := lint.RunAll(lint.Options{
		Path:         input.ProjectPath,
		ChangedFiles: input.ChangedFiles,
	})
	if err != nil {
		return LinterOutput{}, err
	}

	return LinterOutput{
		Passed:       result.Passed,
		ChecksRun:    result.ChecksRun,
		ChecksPassed: result.ChecksPassed,
		ChecksFailed: result.ChecksFailed,
		Findings:     result.Findings,
	}, nil
}
