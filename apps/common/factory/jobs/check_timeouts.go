package jobs

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/jobs"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// CheckTimeoutsJob finds running runs that have exceeded their timeout and fails them.
type CheckTimeoutsJob struct {
	jobs.Base
}

func (j *CheckTimeoutsJob) Name() string { return "factory.check-timeouts" }

func (j *CheckTimeoutsJob) Config() jobs.Config {
	return jobs.Config{
		Queue:   "default",
		Timeout: 30 * time.Second,
	}
}

func (j *CheckTimeoutsJob) Handle(ctx context.Context, _ []byte) error {
	_, err := svc.S.Timeout.Execute(ctx, services.TimeoutInput{})
	return err
}

func (j *CheckTimeoutsJob) Description() string { return "Check and fail tasks that exceeded timeout" }
