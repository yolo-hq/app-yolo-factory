package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// SentinelJob runs the SentinelService and persists findings as tasks/suggestions.
type SentinelJob struct {
	jobs.Base
	ProjectRead     entity.ReadRepository[entities.Project]
	TaskWrite       entity.WriteRepository[entities.Task]
	SuggestionWrite entity.WriteRepository[entities.Suggestion]
}

type sentinelPayload struct {
	ProjectID string   `json:"project_id"`
	Watches   []string `json:"watches"`
}

func (j *SentinelJob) Name() string { return "factory.sentinel" }

func (j *SentinelJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "default",
		MaxRetries: 1,
		Timeout:    5 * time.Minute,
	}
}

func (j *SentinelJob) Handle(ctx context.Context, payload []byte) error {
	var p sentinelPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	// Load project.
	project, err := j.ProjectRead.FindOne(ctx, entity.FindOneOptions{ID: p.ProjectID})
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project %s not found", p.ProjectID)
	}

	// Run sentinel.
	out, err := svc.S.Sentinel.Execute(ctx, services.SentinelInput{
		Project: *project,
		Watches: p.Watches,
	})
	if err != nil {
		return fmt.Errorf("sentinel: %w", err)
	}

	// Persist tasks.
	for i := range out.TasksToCreate {
		if _, err := j.TaskWrite.Insert(ctx, &out.TasksToCreate[i]); err != nil {
			return fmt.Errorf("insert task: %w", err)
		}
	}

	// Persist suggestions.
	for i := range out.SuggestionsToCreate {
		if _, err := j.SuggestionWrite.Insert(ctx, &out.SuggestionsToCreate[i]); err != nil {
			return fmt.Errorf("insert suggestion: %w", err)
		}
	}

	return nil
}

func (j *SentinelJob) Description() string { return "Run sentinel health checks on all active projects" }
