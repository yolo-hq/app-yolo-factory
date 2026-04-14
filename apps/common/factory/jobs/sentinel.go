package jobs

import (
	"context"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// SentinelPayload is the JSON payload for the sentinel job.
type SentinelPayload struct {
	ProjectID string   `json:"project_id"`
	Watches   []string `json:"watches"`
}

// Sentinel runs health checks on active projects.
//
// @name factory.sentinel
// @queue default
// @retries 1
// @timeout 5m
func Sentinel(ctx context.Context, p SentinelPayload) error {
	_, err := svc.S.Sentinel.Execute(ctx, services.SentinelInput{
		ProjectID: p.ProjectID,
		Watches:   p.Watches,
	})
	return err
}
