package jobs

import (
	"context"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// CheckTimeoutsPayload is an empty payload marker.
type CheckTimeoutsPayload struct{}

// CheckTimeouts fails runs that have exceeded their timeout.
//
// @name factory.check-timeouts
// @queue default
// @timeout 30s
func CheckTimeouts(ctx context.Context, _ CheckTimeoutsPayload) error {
	_, err := svc.S.Timeout.Execute(ctx, services.TimeoutInput{})
	return err
}
