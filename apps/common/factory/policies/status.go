// Package policies defines EntityPolicies for Factory domain status guards.
// These replace inline `if status != X` checks in action Execute() methods.
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

// ProjectMustBeActive denies if project status is not "active".
// Used by: SubmitPRDAction, PauseProjectAction
type ProjectMustBeActive struct{ policy.EntityPolicyBase }

func (p *ProjectMustBeActive) PolicyData() any { return &statusData{} }
func (p *ProjectMustBeActive) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.ProjectActive {
		return policy.Deny("project must be active")
	}
	return policy.Allow()
}

// ProjectMustBePaused denies if project status is not "paused".
// Used by: ResumeProjectAction
type ProjectMustBePaused struct{ policy.EntityPolicyBase }

func (p *ProjectMustBePaused) PolicyData() any { return &statusData{} }
func (p *ProjectMustBePaused) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.ProjectPaused {
		return policy.Deny("project must be paused to resume")
	}
	return policy.Allow()
}

// --- PRD policies ---

// PRDMustBeDraft denies if PRD status is not "draft".
// Used by: ApprovePRDAction
type PRDMustBeDraft struct{ policy.EntityPolicyBase }

func (p *PRDMustBeDraft) PolicyData() any { return &statusData{} }
func (p *PRDMustBeDraft) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.PRDDraft {
		return policy.Deny("PRD must be in draft status")
	}
	return policy.Allow()
}

// PRDMustBeDraftOrApproved denies if PRD status is not "draft" or "approved".
// Used by: ExecutePRDAction
type PRDMustBeDraftOrApproved struct{ policy.EntityPolicyBase }

func (p *PRDMustBeDraftOrApproved) PolicyData() any { return &statusData{} }
func (p *PRDMustBeDraftOrApproved) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.PRDDraft && d.Status != entities.PRDApproved {
		return policy.Deny(fmt.Sprintf("PRD must be in draft or approved status, got %q", d.Status))
	}
	return policy.Allow()
}

// --- Task policies ---

// TaskMustBeFailed denies if task status is not "failed".
// Used by: RetryTaskAction
type TaskMustBeFailed struct{ policy.EntityPolicyBase }

func (p *TaskMustBeFailed) PolicyData() any { return &statusData{} }
func (p *TaskMustBeFailed) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.TaskFailed {
		return policy.Deny("task must be failed to retry")
	}
	return policy.Allow()
}

// --- Question policies ---

// QuestionMustBeOpen denies if question status is not "open".
// Used by: AnswerQuestionAction
type QuestionMustBeOpen struct{ policy.EntityPolicyBase }

func (p *QuestionMustBeOpen) PolicyData() any { return &statusData{} }
func (p *QuestionMustBeOpen) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.QuestionOpen {
		return policy.Deny("question must be open to answer")
	}
	return policy.Allow()
}

// --- Insight policies ---

// InsightMustBePending denies if insight status is not "pending".
// Used by: AcknowledgeInsightAction
type InsightMustBePending struct{ policy.EntityPolicyBase }

func (p *InsightMustBePending) PolicyData() any { return &statusData{} }
func (p *InsightMustBePending) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.InsightPending {
		return policy.Deny("insight must be pending")
	}
	return policy.Allow()
}

// InsightMustBeAcknowledged denies if insight status is not "acknowledged".
// Used by: ApplyInsightAction
type InsightMustBeAcknowledged struct{ policy.EntityPolicyBase }

func (p *InsightMustBeAcknowledged) PolicyData() any { return &statusData{} }
func (p *InsightMustBeAcknowledged) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.InsightAcknowledged {
		return policy.Deny("insight must be acknowledged to apply")
	}
	return policy.Allow()
}
