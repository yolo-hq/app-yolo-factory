package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// AdvisorJob runs the AdvisorService and persists suggestions.
type AdvisorJob struct {
	jobs.Base
	Advisor         *services.AdvisorService
	ProjectRead     entity.ReadRepository[entities.Project]
	TaskRead        entity.ReadRepository[entities.Task]
	RunRead         entity.ReadRepository[entities.Run]
	SuggestionWrite entity.WriteRepository[entities.Suggestion]
}

type advisorPayload struct {
	ProjectID    string `json:"project_id"`
	AnalysisType string `json:"analysis_type"`
}

func (j *AdvisorJob) Name() string { return "factory.advisor" }

func (j *AdvisorJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "default",
		MaxRetries: 1,
		Timeout:    10 * time.Minute,
	}
}

func (j *AdvisorJob) Handle(ctx context.Context, payload []byte) error {
	var p advisorPayload
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

	// Load tasks for this project.
	taskResult, err := j.TaskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "project_id", Operator: entity.OpEq, Value: p.ProjectID},
		},
	})
	if err != nil {
		return fmt.Errorf("load tasks: %w", err)
	}

	// Collect task IDs to scope runs to this project.
	taskIDs := make([]any, len(taskResult.Data))
	for i, t := range taskResult.Data {
		taskIDs[i] = t.ID
	}
	if len(taskIDs) == 0 {
		// No tasks means no runs to analyze.
		return nil
	}

	// Load recent runs scoped to this project's tasks.
	runResult, err := j.RunRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "task_id", Operator: entity.OpIn, Value: taskIDs},
		},
		Pagination: &entity.PaginationParams{Limit: 20},
		Sort:       &entity.SortParams{Field: "created_at", Order: "desc"},
	})
	if err != nil {
		return fmt.Errorf("load runs: %w", err)
	}

	// Run advisor.
	out, err := j.Advisor.Execute(ctx, services.AdvisorInput{
		Project:      *project,
		AnalysisType: p.AnalysisType,
		RunHistory:   runResult.Data,
	})
	if err != nil {
		return fmt.Errorf("advisor: %w", err)
	}

	// Persist suggestions.
	for i := range out.Suggestions {
		if _, err := j.SuggestionWrite.Insert(ctx, &out.Suggestions[i]); err != nil {
			return fmt.Errorf("insert suggestion: %w", err)
		}
	}

	return nil
}
