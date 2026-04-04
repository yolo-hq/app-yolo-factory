package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/server/factory/services"
)

// --- Backup ---

type Backup struct {
	command.Base
}

type BackupInput struct {
	StatePath string `flag:"state-path" usage:"Path to backup state directory (default: .factory-state)"`
}

func (c *Backup) Name() string        { return "backup" }
func (c *Backup) Description() string { return "Trigger a manual backup snapshot" }
func (c *Backup) Input() any          { return &BackupInput{} }

func (c *Backup) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*BackupInput)

	statePath := input.StatePath
	if statePath == "" {
		statePath = ".factory-state"
	}

	svc := &services.BackupService{StatePath: statePath}

	cctx.Print("Running backup to %s...", statePath)
	out, err := svc.Execute(ctx, services.BackupInput{
		Trigger:    "manual",
		EntityType: "snapshot",
		EntityID:   "manual",
		EntityData: map[string]string{"trigger": "cli"},
	})
	if err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	cctx.Print("Backup complete: %s (commit %s)", out.FilePath, out.CommitHash)
	cctx.Print("For full entity backup, use the worker (backs up all entity types).")
	return nil
}

// --- Recover ---

type Recover struct {
	command.Base
}

type RecoverInput struct {
	From string `flag:"from" validate:"required" usage:"Path to backup state directory"`
}

func (c *Recover) Name() string        { return "recover" }
func (c *Recover) Description() string { return "Recover factory state from backup" }
func (c *Recover) Input() any          { return &RecoverInput{} }

func (c *Recover) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*RecoverInput)

	svc := &services.BackupService{StatePath: input.From}

	cctx.Print("Recovering from %s...", input.From)
	results, err := svc.Recover(ctx)
	if err != nil {
		return fmt.Errorf("recover: %w", err)
	}

	cctx.Print("Found %d entities in backup.", len(results))
	cctx.Print("NOTE: To restore into the database, use the API with these recovered entities.")
	return nil
}
