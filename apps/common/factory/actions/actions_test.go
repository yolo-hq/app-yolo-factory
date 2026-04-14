package actions

import (
	"testing"

	"github.com/yolo-hq/yolo/core/action"
)

// interface compliance checks — compile-time only.
var (
	_ action.Action = (*CreateProjectAction)(nil)
	_ action.Action = (*UpdateProjectAction)(nil)
	_ action.Action = (*ProjectArchiveAction)(nil)
	_ action.Action = (*ProjectPauseAction)(nil)
	_ action.Action = (*ProjectResumeAction)(nil)
	_ action.Action = (*SubmitPRDAction)(nil)
	_ action.Action = (*PRDApproveAction)(nil)
	_ action.Action = (*ExecutePRDAction)(nil)
	_ action.Action = (*InsightAcknowledgeAction)(nil)
	_ action.Action = (*InsightApplyAction)(nil)
	_ action.Action = (*InsightDismissAction)(nil)
	_ action.Action = (*AnswerQuestionAction)(nil)
	_ action.Action = (*ApproveSuggestionAction)(nil)
	_ action.Action = (*RejectSuggestionAction)(nil)
	_ action.Action = (*TaskCancelAction)(nil)
	_ action.Action = (*RetryTaskAction)(nil)
)


func TestActions_Description(t *testing.T) {
	cases := []struct {
		name   string
		action action.Action
	}{
		{"CreateProject", &CreateProjectAction{}},
		{"UpdateProject", &UpdateProjectAction{}},
		{"ArchiveProject", &ProjectArchiveAction{}},
		{"PauseProject", &ProjectPauseAction{}},
		{"ResumeProject", &ProjectResumeAction{}},
		{"SubmitPRD", &SubmitPRDAction{}},
		{"ApprovePRD", &PRDApproveAction{}},
		{"ExecutePRD", &ExecutePRDAction{}},
		{"AcknowledgeInsight", &InsightAcknowledgeAction{}},
		{"ApplyInsight", &InsightApplyAction{}},
		{"DismissInsight", &InsightDismissAction{}},
		{"AnswerQuestion", &AnswerQuestionAction{}},
		{"ApproveSuggestion", &ApproveSuggestionAction{}},
		{"RejectSuggestion", &RejectSuggestionAction{}},
		{"CancelTask", &TaskCancelAction{}},
		{"RetryTask", &RetryTaskAction{}},
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
