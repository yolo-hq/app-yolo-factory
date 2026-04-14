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

// ProcessAdvisorJob runs the ProcessAdvisorService and persists insights.
type ProcessAdvisorJob struct {
	jobs.Base
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

	_, err := svc.S.ProcessAdvisor.Execute(ctx, services.ProcessAdvisorInput{
		ProjectID: p.ProjectID,
	})
	return err
}

func (j *ProcessAdvisorJob) Description() string {
	return "Analyze factory process metrics and generate insights"
}
