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

// PlanPRDJob runs the PlannerService to plan and create tasks from a PRD.
// Payload fields are on the struct so jobs.Defer(ctx, &PlanPRDJob{PRDID: id}) works.
type PlanPRDJob struct {
	jobs.Base
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
	var p PlanPRDJob
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	_, err := svc.S.Planner.Execute(ctx, services.PlannerInput{PRDID: p.PRDID})
	return err
}

func (j *PlanPRDJob) Description() string { return "Plan and create tasks from an approved PRD" }
