package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/services"
)

// PlanPRDJob runs the PlannerService and persists the resulting tasks.
type PlanPRDJob struct {
	jobs.Base
	Planner     *services.PlannerService
	PRDRead     entity.ReadRepository[entities.PRD]
	PRDWrite    entity.WriteRepository[entities.PRD]
	TaskWrite   entity.WriteRepository[entities.Task]
	ProjectRead entity.ReadRepository[entities.Project]
}

type planPRDPayload struct {
	PRDID string `json:"prd_id"`
}

func (j *PlanPRDJob) Name() string { return "factory.plan-prd" }

func (j *PlanPRDJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "execution",
		MaxRetries: 1,
		Timeout:    10 * time.Minute,
	}
}

func (j *PlanPRDJob) Handle(ctx context.Context, payload []byte) error {
	var p planPRDPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	// Load PRD.
	prd, err := j.PRDRead.FindOne(ctx, entity.FindOneOptions{ID: p.PRDID})
	if err != nil {
		return fmt.Errorf("load prd: %w", err)
	}
	if prd == nil {
		return fmt.Errorf("prd %s not found", p.PRDID)
	}

	// Load Project.
	project, err := j.ProjectRead.FindOne(ctx, entity.FindOneOptions{ID: prd.ProjectID})
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project %s not found", prd.ProjectID)
	}

	// Run planner.
	out, err := j.Planner.Execute(ctx, services.PlannerInput{
		PRD:     *prd,
		Project: *project,
	})
	if err != nil {
		// Mark PRD as failed on planner error.
		_, _ = j.PRDWrite.Update(ctx).
			WhereID(prd.ID).
			Set("status", "failed").
			Exec(ctx)
		return fmt.Errorf("planner: %w", err)
	}

	// Persist tasks.
	for i := range out.Tasks {
		if _, err := j.TaskWrite.Insert(ctx, &out.Tasks[i]); err != nil {
			return fmt.Errorf("insert task %s: %w", out.Tasks[i].Title, err)
		}
	}

	// Update PRD: set total_tasks and transition to approved.
	_, err = j.PRDWrite.Update(ctx).
		WhereID(prd.ID).
		Set("total_tasks", out.Count).
		Set("status", "approved").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update prd: %w", err)
	}

	return nil
}
