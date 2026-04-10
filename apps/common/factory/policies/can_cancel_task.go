package policies

import (
	"context"
	"fmt"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
)

// CanCancelTaskData declares the entity fields this policy reads.
type CanCancelTaskData struct {
	Status string `field:"status"`
}

// CanCancelTaskPolicy denies if task is in a terminal state (done, failed, cancelled).
type CanCancelTaskPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanCancelTaskData]
}

func (p *CanCancelTaskPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	switch data.Status {
	case string(enums.TaskStatusDone), string(enums.TaskStatusFailed), string(enums.TaskStatusCancelled):
		return policy.Deny(fmt.Sprintf("task cannot be cancelled in %q status", data.Status))
	}
	return policy.Allow()
}
