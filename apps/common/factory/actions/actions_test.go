package actions_test

import (
	"testing"

	"github.com/yolo-hq/yolo/core/action"

	actionsgen "github.com/yolo-hq/app-yolo-factory/.yolo/gen/adapters/apps/common/factory/actions"
)

// interface compliance checks — compile-time only.
var (
	_ action.Action = (*actionsgen.ProjectCreateAction)(nil)
	_ action.Action = (*actionsgen.ProjectUpdateAction)(nil)
	_ action.Action = (*actionsgen.ProjectArchiveAction)(nil)
	_ action.Action = (*actionsgen.ProjectPauseAction)(nil)
	_ action.Action = (*actionsgen.ProjectResumeAction)(nil)
	_ action.Action = (*actionsgen.PRDSubmitAction)(nil)
	_ action.Action = (*actionsgen.PRDApproveAction)(nil)
	_ action.Action = (*actionsgen.PRDExecuteAction)(nil)
	_ action.Action = (*actionsgen.InsightAcknowledgeAction)(nil)
	_ action.Action = (*actionsgen.InsightApplyAction)(nil)
	_ action.Action = (*actionsgen.InsightDismissAction)(nil)
	_ action.Action = (*actionsgen.QuestionAnswerAction)(nil)
	_ action.Action = (*actionsgen.SuggestionApproveAction)(nil)
	_ action.Action = (*actionsgen.SuggestionRejectAction)(nil)
	_ action.Action = (*actionsgen.TaskCancelAction)(nil)
	_ action.Action = (*actionsgen.TaskRetryAction)(nil)
	_ action.Action = (*actionsgen.RunCompleteAction)(nil)
)

func TestActions_Description(t *testing.T) {
	cases := []struct {
		name   string
		action action.Action
	}{
		{"ProjectCreate", actionsgen.ProjectCreateAction{}},
		{"ProjectUpdate", actionsgen.ProjectUpdateAction{}},
		{"ProjectArchive", actionsgen.ProjectArchiveAction{}},
		{"ProjectPause", actionsgen.ProjectPauseAction{}},
		{"ProjectResume", actionsgen.ProjectResumeAction{}},
		{"PRDSubmit", actionsgen.PRDSubmitAction{}},
		{"PRDApprove", actionsgen.PRDApproveAction{}},
		{"PRDExecute", actionsgen.PRDExecuteAction{}},
		{"InsightAcknowledge", actionsgen.InsightAcknowledgeAction{}},
		{"InsightApply", actionsgen.InsightApplyAction{}},
		{"InsightDismiss", actionsgen.InsightDismissAction{}},
		{"QuestionAnswer", actionsgen.QuestionAnswerAction{}},
		{"SuggestionApprove", actionsgen.SuggestionApproveAction{}},
		{"SuggestionReject", actionsgen.SuggestionRejectAction{}},
		{"TaskCancel", actionsgen.TaskCancelAction{}},
		{"TaskRetry", actionsgen.TaskRetryAction{}},
		{"RunComplete", actionsgen.RunCompleteAction{}},
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
