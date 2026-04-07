package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestAdvisor_BuildPrompt(t *testing.T) {
	ctxSvc := &ContextService{}
	out, err := ctxSvc.Execute(context.Background(), ContextInput{
		Phase: "advisor",
		Project: entities.Project{
			Name: "my-project",
		},
		AnalysisType:    "code_quality",
		AnalysisContext: "Project has 50 files, 2000 LOC",
		RunHistory:      "- Run 1: status=completed model=sonnet cost=$0.50",
	})
	require.NoError(t, err)

	assertContains(t, out.Prompt, "my-project")
	assertContains(t, out.Prompt, "code_quality")
	assertContains(t, out.Prompt, "Run 1")
	assertContains(t, out.SystemPrompt, "advisor")
}

func TestAdvisor_ParseOutput(t *testing.T) {
	raw := json.RawMessage(`{
		"suggestions": [
			{
				"title": "Extract helper package",
				"body": "Move shared utils to a common package",
				"category": "refactoring",
				"priority": "medium",
				"estimated_impact": "Reduces duplication by 30%"
			}
		]
	}`)

	defs, err := parseAdvisorOutput(raw)
	require.NoError(t, err)
	assert.Len(t, defs, 1)
	assert.Equal(t, "Extract helper package", defs[0].Title)
	assert.Equal(t, "refactoring", defs[0].Category)
}

func TestAdvisor_ParseOutputEmpty(t *testing.T) {
	_, err := parseAdvisorOutput(nil)
	assert.Error(t, err)
}

func TestAdvisor_MapCategory(t *testing.T) {
	assert.Equal(t, "refactor", mapAdvisorCategory("optimization"))
	assert.Equal(t, "refactor", mapAdvisorCategory("refactoring"))
	assert.Equal(t, "refactor", mapAdvisorCategory("tech_debt"))
	assert.Equal(t, "feature", mapAdvisorCategory("new_feature"))
	assert.Equal(t, "feature", mapAdvisorCategory("pattern_extraction"))
	assert.Equal(t, "refactor", mapAdvisorCategory("unknown"))
}

func TestAdvisor_FormatRunHistory(t *testing.T) {
	runs := []entities.Run{
		{Status: "completed", Model: "sonnet", CostUSD: 0.50},
		{Status: "failed", Model: "opus", CostUSD: 1.20, Error: "build failed"},
	}
	runs[0].ID = "run-1"
	runs[1].ID = "run-2"

	result := formatRunHistory(runs)
	assert.Contains(t, result, "run-1")
	assert.Contains(t, result, "completed")
	assert.Contains(t, result, "run-2")
	assert.Contains(t, result, "build failed")
}

func TestAdvisor_FormatRunHistoryEmpty(t *testing.T) {
	result := formatRunHistory(nil)
	assert.Equal(t, "No recent runs.", result)
}
