package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
)

// CleanupRunsInput holds the cleanup-runs command flags.
type CleanupRunsInput struct {
	OlderThan string `json:"older_than" validate:"required"` // duration string: "72h", "7d"
	DryRun    bool   `json:"dry_run"`
}

// CleanupRuns removes old completed/failed runs.
type CleanupRuns struct {
	command.Base
}

func (c *CleanupRuns) Name() string        { return "cleanup-runs" }
func (c *CleanupRuns) Description() string { return "Clean up old completed/failed runs" }
func (c *CleanupRuns) Input() any          { return &CleanupRunsInput{} }

func (c *CleanupRuns) Execute(ctx context.Context, cctx command.Context) error {
	input := cctx.TypedInput.(*CleanupRunsInput)

	cctx.Print("Cleaning up runs older than %s...", input.OlderThan)

	if input.DryRun {
		cctx.Print("(dry run — no changes)")
	}

	// TODO: query runs by age, soft-delete old ones
	repo, err := cctx.RepoProvider.Repo("Run")
	if err != nil {
		return fmt.Errorf("get run repo: %w", err)
	}
	_ = repo

	cctx.Print("Done.")
	return nil
}
