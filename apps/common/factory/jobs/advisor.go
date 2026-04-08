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

// AdvisorJob runs the AdvisorService and persists suggestions.
type AdvisorJob struct {
	jobs.Base
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
	_, err := svc.S.Advisor.Execute(ctx, services.AdvisorInput{
		ProjectID:    p.ProjectID,
		AnalysisType: p.AnalysisType,
	})
	return err
}

func (j *AdvisorJob) Description() string { return "Run optimization advisor on all active projects" }
