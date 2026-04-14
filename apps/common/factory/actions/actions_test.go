package actions

import (
	"testing"

	"github.com/yolo-hq/yolo/core/action"
)

// interface compliance checks — compile-time only.
var (
	_ action.Action = (*ProjectCreateAction)(nil)
	_ action.Action = (*ProjectUpdateAction)(nil)
	_ action.Action = (*ProjectArchiveAction)(nil)
	_ action.Action = (*ProjectPauseAction)(nil)
	_ action.Action = (*ProjectResumeAction)(nil)
	_ action.Action = (*PRDSubmitAction)(nil)
	_ action.Action = (*PRDApproveAction)(nil)
	_ action.Action = (*PRDExecuteAction)(nil)
	_ action.Action = (*InsightAcknowledgeAction)(nil)
	_ action.Action = (*InsightApplyAction)(nil)
	_ action.Action = (*InsightDismissAction)(nil)
	_ action.Action = (*QuestionAnswerAction)(nil)
	_ action.Action = (*SuggestionApproveAction)(nil)
	_ action.Action = (*SuggestionRejectAction)(nil)
	_ action.Action = (*TaskCancelAction)(nil)
	_ action.Action = (*TaskRetryAction)(nil)
)


func TestActions_Description(t *testing.T) {
	cases := []struct {
		name   string
		action action.Action
	}{
		{"CreateProject", &ProjectCreateAction{}},
		{"UpdateProject", &ProjectUpdateAction{}},
		{"ArchiveProject", &ProjectArchiveAction{}},
		{"PauseProject", &ProjectPauseAction{}},
		{"ResumeProject", &ProjectResumeAction{}},
		{"SubmitPRD", &PRDSubmitAction{}},
		{"ApprovePRD", &PRDApproveAction{}},
		{"ExecutePRD", &PRDExecuteAction{}},
		{"AcknowledgeInsight", &InsightAcknowledgeAction{}},
		{"ApplyInsight", &InsightApplyAction{}},
		{"DismissInsight", &InsightDismissAction{}},
		{"AnswerQuestion", &QuestionAnswerAction{}},
		{"ApproveSuggestion", &SuggestionApproveAction{}},
		{"RejectSuggestion", &SuggestionRejectAction{}},
		{"CancelTask", &TaskCancelAction{}},
		{"RetryTask", &TaskRetryAction{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			type describer interface{ Description() string }
			d, ok := tc.action.(describer)
			if !ok {
				t.Fatalf("%T does not implement Description()", tc.action)
			}
			if d.Description() == "" {
				t.Errorf("%T.Description() returned empty string", tc.action)
			}
		})
	}
}
