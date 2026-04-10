package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestIntegrationReview_BuildPrompt(t *testing.T) {
	ctxSvc := &ContextService{}
	out, err := ctxSvc.Execute(context.Background(), ContextInput{
		Phase: "integration_review",
		Project: entities.Project{
			Name: "test-project",
		},
		TaskSummaries: "- [T1] Add users: create user entity\n- [T2] Add roles: create role entity",
		GitDiff:       "diff --git a/user.go b/user.go\n+func CreateUser() {}",
	})
	require.NoError(t, err)

	assertContains(t, out.Prompt, "test-project")
	assertContains(t, out.Prompt, "Add users")
	assertContains(t, out.Prompt, "Add roles")
	assertContains(t, out.Prompt, "CreateUser")
	assertContains(t, out.SystemPrompt, "integration reviewer")
}

func TestIntegrationReview_ParseFindings(t *testing.T) {
	raw := json.RawMessage(`{
		"findings": [
			{
				"category": "duplicate_function",
				"severity": "warning",
				"message": "formatName() exists in both user.go and role.go",
				"files": ["user.go", "role.go"]
			},
			{
				"category": "inconsistent_pattern",
				"severity": "error",
				"message": "user.go uses string literals, role.go uses constants",
				"files": ["user.go"]
			}
		]
	}`)

	findings, err := parseIntegrationReviewOutput(raw)
	require.NoError(t, err)
	assert.Len(t, findings, 2)
	assert.Equal(t, "duplicate_function", findings[0].Category)
	assert.Equal(t, "warning", findings[0].Severity)
	assert.Equal(t, []string{"user.go", "role.go"}, findings[0].Files)
	assert.Equal(t, "error", findings[1].Severity)
}

func TestIntegrationReview_ParseFindingsEmpty(t *testing.T) {
	_, err := parseIntegrationReviewOutput(nil)
	assert.Error(t, err)
}

func TestIntegrationReview_FindingsToSuggestions(t *testing.T) {
	findings := []IntegrationFinding{
		{Category: "duplicate_function", Severity: "warning", Message: "dup func"},
		{Category: "inconsistent_pattern", Severity: "error", Message: "bad pattern"},
		{Category: "missing_helper", Severity: "info", Message: "could add helper"},
	}

	suggestions := findingsToSuggestions(findings, "proj-1")
	// info severity should be skipped
	assert.Len(t, suggestions, 2)

	assert.Equal(t, "proj-1", suggestions[0].ProjectID)
	assert.Equal(t, "integration_review", suggestions[0].Source)
	assert.Equal(t, string(enums.SuggestionCategoryRefactoring), suggestions[0].Category)
	assert.Equal(t, string(enums.SuggestionPriorityMedium), suggestions[0].Priority)

	assert.Equal(t, string(enums.SuggestionCategoryTechDebt), suggestions[1].Category)
	assert.Equal(t, string(enums.SuggestionPriorityHigh), suggestions[1].Priority)
}

func TestIntegrationReview_ShouldRun(t *testing.T) {
	assert.False(t, ShouldRunIntegrationReview(0, 5))
	assert.False(t, ShouldRunIntegrationReview(3, 5))
	assert.True(t, ShouldRunIntegrationReview(5, 5))
	assert.True(t, ShouldRunIntegrationReview(10, 5))
	assert.False(t, ShouldRunIntegrationReview(7, 5))

	// default interval
	assert.True(t, ShouldRunIntegrationReview(5, 0))
	assert.True(t, ShouldRunIntegrationReview(10, -1))
}

func TestIntegrationReview_FormatTaskSummaries(t *testing.T) {
	tasks := []entities.Task{
		{Title: "Add users", Spec: "Create user entity"},
		{Title: "Add roles", Spec: "Create role entity"},
	}
	tasks[0].ID = "t1"
	tasks[1].ID = "t2"

	result := formatTaskSummaries(tasks)
	assert.Contains(t, result, "t1")
	assert.Contains(t, result, "Add users")
	assert.Contains(t, result, "t2")
	assert.Contains(t, result, "Add roles")
}

func TestIntegrationReview_FormatTaskSummariesEmpty(t *testing.T) {
	result := formatTaskSummaries(nil)
	assert.Equal(t, "No recent tasks.", result)
}

func TestIntegrationReview_MapFindingCategory(t *testing.T) {
	assert.Equal(t, string(enums.SuggestionCategoryRefactoring), mapFindingCategory("duplicate_function"))
	assert.Equal(t, string(enums.SuggestionCategoryRefactoring), mapFindingCategory("missing_helper"))
	assert.Equal(t, string(enums.SuggestionCategoryTechDebt), mapFindingCategory("inconsistent_pattern"))
	assert.Equal(t, string(enums.SuggestionCategoryTechDebt), mapFindingCategory("state_machine_drift"))
	assert.Equal(t, string(enums.SuggestionCategoryRefactoring), mapFindingCategory("unknown"))
}
