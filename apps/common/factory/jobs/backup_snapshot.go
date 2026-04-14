package jobs

import (
	"context"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// BackupSnapshotPayload is an empty payload marker.
type BackupSnapshotPayload struct{}

// BackupSnapshot creates periodic backup snapshots of factory state.
//
// @name factory.backup-snapshot
// @queue default
// @retries 1
// @timeout 5m
func BackupSnapshot(ctx context.Context, _ BackupSnapshotPayload) error {
	_, err := svc.S.Backup.Execute(ctx, services.BackupInput{Trigger: "snapshot"})
	return err
}
