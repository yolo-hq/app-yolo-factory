package jobs

import (
	"context"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// AdvisorPayload is the JSON payload for the advisor job.
type AdvisorPayload struct {
	ProjectID    string `json:"project_id"`
	AnalysisType string `json:"analysis_type"`
}

// Advisor runs optimization advisor on active projects.
//
// @name factory.advisor
// @queue default
// @retries 1
// @timeout 10m
func Advisor(ctx context.Context, p AdvisorPayload) error {
	_, err := svc.S.Advisor.Execute(ctx, services.AdvisorInput{
		ProjectID:    p.ProjectID,
		AnalysisType: p.AnalysisType,
	})
	return err
}
