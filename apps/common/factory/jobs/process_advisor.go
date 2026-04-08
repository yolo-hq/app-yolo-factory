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

// ProcessAdvisorJob runs the ProcessAdvisorService and persists insights.
type ProcessAdvisorJob struct {
	jobs.Base
	ProjectRead    entity.ReadRepository[entities.Project]
	TaskRead       entity.ReadRepository[entities.Task]
	RunRead        entity.ReadRepository[entities.Run]
	StepRead       entity.ReadRepository[entities.Step]
	ReviewRead     entity.ReadRepository[entities.Review]
	LintResultRead entity.ReadRepository[entities.LintResult]
	InsightWrite   entity.WriteRepository[entities.Insight]
}

type processAdvisorPayload struct {
	ProjectID string `json:"project_id"`
}

func (j *ProcessAdvisorJob) Name() string { return "factory.process-advisor" }

func (j *ProcessAdvisorJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "default",
		MaxRetries: 1,
		Timeout:    10 * time.Minute,
	}
}

func (j *ProcessAdvisorJob) Handle(ctx context.Context, payload []byte) error {
	var p processAdvisorPayload
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

	// Load all tasks for project.
	taskResult, err := j.TaskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "project_id", Operator: entity.OpEq, Value: p.ProjectID},
		},
	})
	if err != nil {
		return fmt.Errorf("load tasks: %w", err)
	}

	// Load all runs.
	runResult, err := j.RunRead.FindMany(ctx, entity.FindOptions{
		Pagination: &entity.PaginationParams{Limit: 500},
		Sort:       &entity.SortParams{Field: "created_at", Order: "desc"},
	})
	if err != nil {
		return fmt.Errorf("load runs: %w", err)
	}

	// Load all steps.
	stepResult, err := j.StepRead.FindMany(ctx, entity.FindOptions{
		Pagination: &entity.PaginationParams{Limit: 1000},
	})
	if err != nil {
		return fmt.Errorf("load steps: %w", err)
	}

	// Load reviews.
	reviewResult, err := j.ReviewRead.FindMany(ctx, entity.FindOptions{
		Pagination: &entity.PaginationParams{Limit: 500},
	})
	if err != nil {
		return fmt.Errorf("load reviews: %w", err)
	}

	// Load lint results.
	lintResult, err := j.LintResultRead.FindMany(ctx, entity.FindOptions{
		Pagination: &entity.PaginationParams{Limit: 500},
	})
	if err != nil {
		return fmt.Errorf("load lint results: %w", err)
	}

	// Run process advisor.
	out, err := svc.S.ProcessAdvisor.Execute(ctx, services.ProcessAdvisorInput{
		ProjectID:   p.ProjectID,
		Tasks:       taskResult.Data,
		Runs:        runResult.Data,
		Steps:       stepResult.Data,
		Reviews:     reviewResult.Data,
		LintResults: lintResult.Data,
	})
	if err != nil {
		return fmt.Errorf("process advisor: %w", err)
	}

	if out.Skipped {
		return nil
	}

	// Persist insights.
	for i := range out.Insights {
		if _, err := j.InsightWrite.Insert(ctx, &out.Insights[i]); err != nil {
			return fmt.Errorf("insert insight: %w", err)
		}
	}

	return nil
}

func (j *ProcessAdvisorJob) Description() string { return "Analyze factory process metrics and generate insights" }
