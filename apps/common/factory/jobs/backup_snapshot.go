package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// BackupSnapshotJob dumps all entities to the backup state repo.
type BackupSnapshotJob struct {
	jobs.Base
	Backup          *services.BackupService
	ProjectRead     entity.ReadRepository[entities.Project]
	PRDRead         entity.ReadRepository[entities.PRD]
	TaskRead        entity.ReadRepository[entities.Task]
	QuestionRead    entity.ReadRepository[entities.Question]
	SuggestionRead  entity.ReadRepository[entities.Suggestion]
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
	if j.Backup == nil {
		return fmt.Errorf("factory.backup-snapshot: dependencies not injected")
	}

	// Back up each entity type.
	if err := j.backupAll(ctx, "project", j.ProjectRead); err != nil {
		return fmt.Errorf("backup projects: %w", err)
	}
	if err := j.backupAll(ctx, "prd", j.PRDRead); err != nil {
		return fmt.Errorf("backup prds: %w", err)
	}
	if err := j.backupAll(ctx, "task", j.TaskRead); err != nil {
		return fmt.Errorf("backup tasks: %w", err)
	}
	if err := j.backupAll(ctx, "question", j.QuestionRead); err != nil {
		return fmt.Errorf("backup questions: %w", err)
	}
	if err := j.backupAll(ctx, "suggestion", j.SuggestionRead); err != nil {
		return fmt.Errorf("backup suggestions: %w", err)
	}

	return nil
}

// backupAll is a generic helper that backs up all entities of a given type.
func backupAll[T entity.Entity](ctx context.Context, entityType string, repo entity.ReadRepository[T], backup *services.BackupService) error {
	result, err := repo.FindMany(ctx, entity.FindOptions{})
	if err != nil {
		return err
	}

	for _, e := range result.Data {
		_, err := backup.Execute(ctx, services.BackupInput{
			Trigger:    "daily_snapshot",
			EntityType: entityType,
			EntityID:   e.GetID(),
			EntityData: e,
		})
		if err != nil {
			return fmt.Errorf("backup %s %s: %w", entityType, e.GetID(), err)
		}
	}
	return nil
}

func (j *BackupSnapshotJob) backupAll(ctx context.Context, entityType string, repo any) error {
	switch r := repo.(type) {
	case entity.ReadRepository[entities.Project]:
		return backupAll(ctx, entityType, r, j.Backup)
	case entity.ReadRepository[entities.PRD]:
		return backupAll(ctx, entityType, r, j.Backup)
	case entity.ReadRepository[entities.Task]:
		return backupAll(ctx, entityType, r, j.Backup)
	case entity.ReadRepository[entities.Question]:
		return backupAll(ctx, entityType, r, j.Backup)
	case entity.ReadRepository[entities.Suggestion]:
		return backupAll(ctx, entityType, r, j.Backup)
	default:
		return fmt.Errorf("unknown repo type for %s", entityType)
	}
}
