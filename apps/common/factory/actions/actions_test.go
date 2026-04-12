package actions

import (
	"testing"

	"github.com/yolo-hq/yolo/core/action"
)

// interface compliance checks — compile-time only.
var (
	_ action.Action = (*CreateProjectAction)(nil)
	_ action.Action = (*UpdateProjectAction)(nil)
	_ action.Action = (*ArchiveProjectAction)(nil)
	_ action.Action = (*PauseProjectAction)(nil)
	_ action.Action = (*ResumeProjectAction)(nil)
	_ action.Action = (*SubmitPRDAction)(nil)
	_ action.Action = (*ApprovePRDAction)(nil)
	_ action.Action = (*ExecutePRDAction)(nil)
	_ action.Action = (*AcknowledgeInsightAction)(nil)
	_ action.Action = (*ApplyInsightAction)(nil)
	_ action.Action = (*DismissInsightAction)(nil)
	_ action.Action = (*AnswerQuestionAction)(nil)
	_ action.Action = (*ApproveSuggestionAction)(nil)
	_ action.Action = (*RejectSuggestionAction)(nil)
	_ action.Action = (*CancelTaskAction)(nil)
	_ action.Action = (*RetryTaskAction)(nil)
)


func TestActions_Description(t *testing.T) {
	cases := []struct {
		name   string
		action action.Action
	}{
		{"CreateProject", &CreateProjectAction{}},
		{"UpdateProject", &UpdateProjectAction{}},
		{"ArchiveProject", &ArchiveProjectAction{}},
		{"PauseProject", &PauseProjectAction{}},
		{"ResumeProject", &ResumeProjectAction{}},
		{"SubmitPRD", &SubmitPRDAction{}},
		{"ApprovePRD", &ApprovePRDAction{}},
		{"ExecutePRD", &ExecutePRDAction{}},
		{"AcknowledgeInsight", &AcknowledgeInsightAction{}},
		{"ApplyInsight", &ApplyInsightAction{}},
		{"DismissInsight", &DismissInsightAction{}},
		{"AnswerQuestion", &AnswerQuestionAction{}},
		{"ApproveSuggestion", &ApproveSuggestionAction{}},
		{"RejectSuggestion", &RejectSuggestionAction{}},
		{"CancelTask", &CancelTaskAction{}},
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
