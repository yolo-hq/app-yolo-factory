package commands

import (
	"context"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// BackupInput is the CLI input for the backup command.
type BackupInput struct {
	StatePath string `flag:"state-path" usage:"Path to backup state directory (default: .factory-state)"`
}

// Backup triggers a manual backup snapshot.
//
// @name backup
func Backup(ctx context.Context, cctx *command.Context, in BackupInput) error {
	statePath := in.StatePath
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
