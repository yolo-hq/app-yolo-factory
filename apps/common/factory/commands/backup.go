package commands

import (
	"context"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

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
		return err
	}

	cctx.Print("Backup complete: %s (commit %s)", out.FilePath, out.CommitHash)
	cctx.Print("For full entity backup, use the worker (backs up all entity types).")
	return nil
}
