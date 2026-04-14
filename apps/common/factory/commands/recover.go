package commands

import (
	"context"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// RecoverInput is the CLI input for the recover command.
type RecoverInput struct {
	From string `flag:"from" validate:"required" usage:"Path to backup state directory"`
}

// Recover restores factory state from a backup.
//
// @name recover
func Recover(ctx context.Context, cctx *command.Context, in RecoverInput) error {
	svc := &services.BackupService{StatePath: in.From}

	cctx.Print("Recovering from %s...", in.From)
	results, err := svc.Recover(ctx)
	if err != nil {
		return err
	}

	cctx.Print("Found %d entities in backup.", len(results))
	cctx.Print("NOTE: To restore into the database, use the API with these recovered entities.")
	return nil
}
