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

// SentinelJob runs the SentinelService for health checks on a project.
type SentinelJob struct {
	jobs.Base
}

type sentinelPayload struct {
	ProjectID string   `json:"project_id"`
	Watches   []string `json:"watches"`
}

func (j *SentinelJob) Name() string { return "factory.sentinel" }

func (j *SentinelJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "default",
		MaxRetries: 1,
		Timeout:    5 * time.Minute,
	}
}

func (j *SentinelJob) Handle(ctx context.Context, payload []byte) error {
	var p sentinelPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}
	_, err := svc.S.Sentinel.Execute(ctx, services.SentinelInput{
		ProjectID: p.ProjectID,
		Watches:   p.Watches,
	})
	return err
}

func (j *SentinelJob) Description() string {
	return "Run sentinel health checks on all active projects"
}
