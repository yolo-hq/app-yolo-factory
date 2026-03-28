package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
)

// RetryFailedInput holds the retry-failed command flags.
type RetryFailedInput struct {
	TaskID string `json:"task_id"` // optional: filter by task
	DryRun bool   `json:"dry_run"`
}

// RetryFailed retries all failed runs.
type RetryFailed struct {
	command.Base
}

func (c *RetryFailed) Name() string        { return "retry-failed" }
func (c *RetryFailed) Description() string { return "Retry all failed runs" }
func (c *RetryFailed) Input() any          { return &RetryFailedInput{} }

func (c *RetryFailed) Execute(ctx context.Context, cctx command.Context) error {
	input := cctx.TypedInput.(*RetryFailedInput)

	if input.TaskID != "" {
		cctx.Print("Retrying failed runs for task %s...", input.TaskID)
	} else {
		cctx.Print("Retrying all failed runs...")
	}

	if input.DryRun {
		cctx.Print("(dry run — no changes)")
	}

	// TODO: query failed runs via repo and re-enqueue execute-task jobs
	repo, err := cctx.RepoProvider.Repo("Run")
	if err != nil {
		return fmt.Errorf("get run repo: %w", err)
	}
	_ = repo

	cctx.Print("Done.")
	return nil
}
