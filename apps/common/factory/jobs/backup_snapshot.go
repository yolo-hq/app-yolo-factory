package jobs

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/jobs"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// BackupSnapshotJob dumps all entities to the backup state repo.
type BackupSnapshotJob struct {
	jobs.Base
}

func (j *BackupSnapshotJob) Name() string { return "factory.backup-snapshot" }

func (j *BackupSnapshotJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "default",
		MaxRetries: 1,
		Timeout:    5 * time.Minute,
	}
}

func (j *BackupSnapshotJob) Handle(ctx context.Context, _ []byte) error {
	_, err := svc.S.Backup.Execute(ctx, services.BackupInput{Trigger: "snapshot"})
	return err
}

func (j *BackupSnapshotJob) Description() string {
	return "Create periodic backup snapshots of factory state"
}
