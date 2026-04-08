// Package policies defines EntityPolicies for Factory domain status guards.
// Named Can{Action}{Entity}Policy — each policy maps 1:1 with an action.
package policies

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// statusData is the policy data struct — requests only the "status" field.
type statusData struct {
	Status string `policy:"status"`
}

// --- Project policies ---

// CanPauseProjectPolicy denies if project status is not "active".
type CanPauseProjectPolicy struct{ policy.EntityPolicyBase }

func (p *CanPauseProjectPolicy) PolicyData() any { return &statusData{} }
func (p *CanPauseProjectPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.ProjectActive {
		return policy.Deny("project must be active")
	}
	return policy.Allow()
}

// CanResumeProjectPolicy denies if project status is not "paused".
type CanResumeProjectPolicy struct{ policy.EntityPolicyBase }

func (p *CanResumeProjectPolicy) PolicyData() any { return &statusData{} }
func (p *CanResumeProjectPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.ProjectPaused {
		return policy.Deny("project must be paused to resume")
	}
	return policy.Allow()
}

// --- PRD policies ---

// CanApprovePRDPolicy denies if PRD status is not "draft".
type CanApprovePRDPolicy struct{ policy.EntityPolicyBase }

func (p *CanApprovePRDPolicy) PolicyData() any { return &statusData{} }
func (p *CanApprovePRDPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.PRDDraft {
		return policy.Deny("PRD must be in draft status")
	}
	return policy.Allow()
}

// CanExecutePRDPolicy denies if PRD status is not "draft" or "approved".
type CanExecutePRDPolicy struct{ policy.EntityPolicyBase }

func (p *CanExecutePRDPolicy) PolicyData() any { return &statusData{} }
func (p *CanExecutePRDPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.PRDDraft && d.Status != entities.PRDApproved {
		return policy.Deny(fmt.Sprintf("PRD must be in draft or approved status, got %q", d.Status))
	}
	return policy.Allow()
}

// --- Task policies ---

// CanRetryTaskPolicy denies if task status is not "failed".
type CanRetryTaskPolicy struct{ policy.EntityPolicyBase }

func (p *CanRetryTaskPolicy) PolicyData() any { return &statusData{} }
func (p *CanRetryTaskPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.TaskFailed {
		return policy.Deny("task must be failed to retry")
	}
	return policy.Allow()
}

// CanCancelTaskPolicy denies if task is in a terminal state (done, failed, cancelled).
type CanCancelTaskPolicy struct{ policy.EntityPolicyBase }

func (p *CanCancelTaskPolicy) PolicyData() any { return &statusData{} }
func (p *CanCancelTaskPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	switch d.Status {
	case entities.TaskDone, entities.TaskFailed, entities.TaskCancelled:
		return policy.Deny(fmt.Sprintf("task cannot be cancelled in %q status", d.Status))
	}
	return policy.Allow()
}

// --- Question policies ---

// CanAnswerQuestionPolicy denies if question status is not "open".
type CanAnswerQuestionPolicy struct{ policy.EntityPolicyBase }

func (p *CanAnswerQuestionPolicy) PolicyData() any { return &statusData{} }
func (p *CanAnswerQuestionPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.QuestionOpen {
		return policy.Deny("question must be open to answer")
	}
	return policy.Allow()
}

// --- Insight policies ---

// CanAcknowledgeInsightPolicy denies if insight status is not "pending".
type CanAcknowledgeInsightPolicy struct{ policy.EntityPolicyBase }

func (p *CanAcknowledgeInsightPolicy) PolicyData() any { return &statusData{} }
func (p *CanAcknowledgeInsightPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.InsightPending {
		return policy.Deny("insight must be pending")
	}
	return policy.Allow()
}

// CanApplyInsightPolicy denies if insight status is not "acknowledged".
type CanApplyInsightPolicy struct{ policy.EntityPolicyBase }

func (p *CanApplyInsightPolicy) PolicyData() any { return &statusData{} }
func (p *CanApplyInsightPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.InsightAcknowledged {
		return policy.Deny("insight must be acknowledged to apply")
	}
	return policy.Allow()
}

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

// --- Suggestion policies ---

// CanApproveSuggestionPolicy denies if suggestion status is not "pending".
type CanApproveSuggestionPolicy struct{ policy.EntityPolicyBase }

func (p *CanApproveSuggestionPolicy) PolicyData() any { return &statusData{} }
func (p *CanApproveSuggestionPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.SuggestionPending {
		return policy.Deny("suggestion must be pending to approve")
	}
	return policy.Allow()
}

// CanRejectSuggestionPolicy denies if suggestion status is not "pending".
type CanRejectSuggestionPolicy struct{ policy.EntityPolicyBase }

func (p *CanRejectSuggestionPolicy) PolicyData() any { return &statusData{} }
func (p *CanRejectSuggestionPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.SuggestionPending {
		return policy.Deny("suggestion must be pending to reject")
	}
	return policy.Allow()
}
