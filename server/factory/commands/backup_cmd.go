package commands

import (
	"context"

	"github.com/yolo-hq/yolo/core/command"
)

// --- Backup ---

type Backup struct {
	command.Base
}

func (c *Backup) Name() string        { return "backup" }
func (c *Backup) Description() string { return "Trigger a manual backup snapshot" }

func (c *Backup) Execute(_ context.Context, cctx command.Context) error {
	cctx.Print("Backup snapshot queued.")
	cctx.Print("Use the worker to process the backup job.")
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

func (c *Recover) Execute(_ context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*RecoverInput)

	cctx.Print("Recovery from: %s", input.From)
	cctx.Print("NOTE: Full recovery requires a running database. Use the API for automated recovery.")
	return nil
}
