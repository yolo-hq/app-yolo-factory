package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/jobs"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// ExecuteWorkflowJob runs the OrchestratorService for a single task.
// Payload fields are on the struct so jobs.Defer(ctx, &ExecuteWorkflowJob{TaskID: id}) works.
type ExecuteWorkflowJob struct {
	jobs.Base
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
	var p ExecuteWorkflowJob
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	_, err := svc.S.Orchestrator.Execute(ctx, services.OrchestratorInput{TaskID: p.TaskID})
	return err
}

func (j *ExecuteWorkflowJob) Description() string {
	return "Run the full implementation workflow for a queued task"
}
