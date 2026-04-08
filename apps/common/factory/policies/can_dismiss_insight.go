package policies

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanDismissInsightPolicy denies if insight is not "pending" or "acknowledged".
type CanDismissInsightPolicy struct{ policy.EntityPolicyBase }

func (p *CanDismissInsightPolicy) PolicyData() any { return &statusData{} }
func (p *CanDismissInsightPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.InsightPending && d.Status != entities.InsightAcknowledged {
		return policy.Deny(fmt.Sprintf("insight must be pending or acknowledged to dismiss, got %q", d.Status))
	}
	return policy.Allow()
}
