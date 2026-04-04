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

// ExecuteWorkflowJob runs the OrchestratorService for a single task and persists all results.
type ExecuteWorkflowJob struct {
	jobs.Base
	Orchestrator *services.OrchestratorService
	TaskRead     entity.ReadRepository[entities.Task]
	TaskWrite    entity.WriteRepository[entities.Task]
	RunWrite     entity.WriteRepository[entities.Run]
	StepWrite    entity.WriteRepository[entities.Step]
	ReviewWrite  entity.WriteRepository[entities.Review]
	PRDRead      entity.ReadRepository[entities.PRD]
	ProjectRead  entity.ReadRepository[entities.Project]
}

type executeWorkflowPayload struct {
	TaskID string `json:"task_id"`
}

func (j *ExecuteWorkflowJob) Name() string { return "factory.execute-workflow" }

func (j *ExecuteWorkflowJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "execution",
		MaxRetries: 1,
		Timeout:    30 * time.Minute,
	}
}

func (j *ExecuteWorkflowJob) Handle(ctx context.Context, payload []byte) error {
	// 1. Parse payload.
	var p executeWorkflowPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	// 2. Load task, PRD, project.
	task, err := j.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: p.TaskID})
	if err != nil {
		return fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task %s not found", p.TaskID)
	}

	prd, err := j.PRDRead.FindOne(ctx, entity.FindOneOptions{ID: task.PrdID})
	if err != nil {
		return fmt.Errorf("load prd: %w", err)
	}
	if prd == nil {
		return fmt.Errorf("prd %s not found", task.PrdID)
	}

	project, err := j.ProjectRead.FindOne(ctx, entity.FindOneOptions{ID: task.ProjectID})
	if err != nil {
		return fmt.Errorf("load project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project %s not found", task.ProjectID)
	}

	// 3. Update task: status -> "running", started_at.
	now := time.Now()
	_, err = j.TaskWrite.Update(ctx).
		WhereID(task.ID).
		Set("status", entities.TaskRunning).
		Set("started_at", now).
		Set("run_count", task.RunCount+1).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update task to running: %w", err)
	}

	// 4. Execute the full workflow.
	out, err := j.Orchestrator.Execute(ctx, services.OrchestratorInput{
		Task:    *task,
		PRD:     *prd,
		Project: *project,
	})
	if err != nil {
		// Orchestrator returned a hard error (not a step failure).
		// No Run entity was created, so mark task failed directly as a safety net.
		if _, taskErr := j.TaskWrite.Update(ctx).
			WhereID(task.ID).
			Set("status", entities.TaskFailed).
			Exec(ctx); taskErr != nil {
			return fmt.Errorf("orchestrator: %w (also failed to update task: %v)", err, taskErr)
		}
		return fmt.Errorf("orchestrator: %w", err)
	}

	// 5. Persist Run.
	if _, err := j.RunWrite.Insert(ctx, &out.Run); err != nil {
		return fmt.Errorf("insert run: %w", err)
	}

	// 6. Persist all Steps.
	for i := range out.Steps {
		if _, err := j.StepWrite.Insert(ctx, &out.Steps[i]); err != nil {
			return fmt.Errorf("insert step %s: %w", out.Steps[i].Phase, err)
		}
	}

	// 7. Persist Review if present.
	if out.Review != nil {
		if _, err := j.ReviewWrite.Insert(ctx, out.Review); err != nil {
			return fmt.Errorf("insert review: %w", err)
		}
	}

	// 8. Task/PRD state transitions are handled by CompleteRunAction,
	// which is triggered on the Run entity after persistence.
	// This job only persists orchestrator output — no duplicate state machine here.

	return nil
}
