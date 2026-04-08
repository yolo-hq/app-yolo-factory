package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

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
