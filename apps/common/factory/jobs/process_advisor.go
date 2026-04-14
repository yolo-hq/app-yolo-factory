package jobs

import (
	"context"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// ProcessAdvisorPayload is the JSON payload for the process-advisor job.
type ProcessAdvisorPayload struct {
	ProjectID string `json:"project_id"`
}

// ProcessAdvisor analyzes factory process metrics and generates insights.
//
// @name factory.process-advisor
// @queue default
// @retries 1
// @timeout 10m
func ProcessAdvisor(ctx context.Context, p ProcessAdvisorPayload) error {
	_, err := svc.S.ProcessAdvisor.Execute(ctx, services.ProcessAdvisorInput{
		ProjectID: p.ProjectID,
	})
	return err
}
