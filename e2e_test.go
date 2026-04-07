package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/core/pkg/claude"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

func testdataPath(name string) string {
	return filepath.Join("testdata", name)
}

func runMockScript(t *testing.T, name string) []byte {
	t.Helper()
	path := testdataPath(name)
	out, err := exec.Command("bash", path).Output()
	require.NoError(t, err, "mock script %s should execute", name)
	return out
}

// TestE2E_MockScripts verifies each mock script returns valid JSON that parses into claude.Result.
func TestE2E_MockScripts(t *testing.T) {
	scripts := []string{
		"mock_claude_planner.sh",
		"mock_claude_implementer.sh",
		"mock_claude_reviewer_pass.sh",
		"mock_claude_reviewer_fail.sh",
		"mock_claude_auditor_pass.sh",
		"mock_claude_question.sh",
	}

	for _, script := range scripts {
		t.Run(script, func(t *testing.T) {
			out := runMockScript(t, script)

			var result claude.Result
			err := json.Unmarshal(out, &result)
			require.NoError(t, err, "output should be valid claude.Result JSON")

			assert.NotEmpty(t, result.SessionID, "session_id should be set")
			assert.False(t, result.IsError, "mock should not be an error result")
			assert.Greater(t, result.CostUSD, 0.0, "cost should be positive")
		})
	}
}

// TestE2E_PlannerParsesOutput verifies PlannerService can parse planner mock output.
func TestE2E_PlannerParsesOutput(t *testing.T) {
	out := runMockScript(t, "mock_claude_planner.sh")

	var result claude.Result
	require.NoError(t, json.Unmarshal(out, &result))

	// The planner uses parseTaskDefs which is internal to services package.
	// Verify the structured_output has expected shape.
	var planOutput struct {
		Tasks []struct {
			Title              string `json:"title"`
			Spec               string `json:"spec"`
			Sequence           int    `json:"sequence"`
			DependsOn          []int  `json:"depends_on"`
			AcceptanceCriteria []struct {
				ID          string `json:"id"`
				Description string `json:"description"`
			} `json:"acceptance_criteria"`
		} `json:"tasks"`
	}
	require.NoError(t, json.Unmarshal(result.StructuredOutput, &planOutput))
	assert.Len(t, planOutput.Tasks, 3)

	assert.Equal(t, "Setup database schema", planOutput.Tasks[0].Title)
	assert.Equal(t, 1, planOutput.Tasks[0].Sequence)
	assert.Empty(t, planOutput.Tasks[0].DependsOn)
	assert.Len(t, planOutput.Tasks[0].AcceptanceCriteria, 2)

	assert.Equal(t, "Implement user CRUD", planOutput.Tasks[1].Title)
	assert.Equal(t, []int{1}, planOutput.Tasks[1].DependsOn)

	assert.Equal(t, "Add user tests", planOutput.Tasks[2].Title)
	assert.Equal(t, []int{2}, planOutput.Tasks[2].DependsOn)
}

// TestE2E_OrchestratorHelpers verifies orchestrator helper functions work correctly.
func TestE2E_OrchestratorHelpers(t *testing.T) {
	// Test determineModel: task > project > fallback.
	project := entities.Project{DefaultModel: "sonnet"}
	task := entities.Task{Model: "opus"}
	assert.Equal(t, "opus", services.ExportDetermineModel(task, project))

	task.Model = ""
	assert.Equal(t, "sonnet", services.ExportDetermineModel(task, project))

	project.DefaultModel = ""
	assert.Equal(t, "sonnet", services.ExportDetermineModel(task, project))

	// Test workingDir.
	project.LocalPath = "/tmp/myproject"
	project.UseWorktrees = false
	assert.Equal(t, "/tmp/myproject", services.ExportWorkingDir(project, "task-1"))

	project.UseWorktrees = true
	assert.Equal(t, "/tmp/myproject/.worktrees/task-task-1", services.ExportWorkingDir(project, "task-1"))

	// Test parseTestCommands.
	assert.Equal(t, []string{"make build"}, services.ExportParseTestCommands(`["make build"]`))
	assert.Equal(t, []string{"go build ./...", "go test ./..."}, services.ExportParseTestCommands(""))

	// Test Truncate.
	assert.Equal(t, "hello", services.Truncate("hello", 100))
	assert.Equal(t, "hello...", services.Truncate("hello world", 5))
}

// TestE2E_DependencyChain verifies cycle detection and dependency parsing.
func TestE2E_DependencyChain(t *testing.T) {
	// Linear chain: A -> B -> C (no cycles).
	tasks := map[string]*entities.Task{
		"A": {DependsOn: "[]"},
		"B": {DependsOn: `["A"]`},
	}
	err := services.ExportDetectCycle("C", []string{"B"}, tasks)
	assert.NoError(t, err)

	// Self-dependency.
	err = services.ExportDetectCycle("A", []string{"A"}, map[string]*entities.Task{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")

	// Three-node cycle: A -> B -> C -> A.
	cycleMap := map[string]*entities.Task{
		"A": {DependsOn: `["B"]`},
		"B": {DependsOn: `["C"]`},
	}
	err = services.ExportDetectCycle("C", []string{"A"}, cycleMap)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")

	// ParseDeps.
	assert.Nil(t, services.ParseDeps(""))
	assert.Nil(t, services.ParseDeps("[]"))
	assert.Equal(t, []string{"a", "b"}, services.ParseDeps(`["a","b"]`))
}

// TestE2E_BackupRoundTrip tests backup serialization and recovery.
func TestE2E_BackupRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "factory-state")

	svc := &services.BackupService{StatePath: statePath}

	// Create a test entity to back up.
	project := entities.Project{
		Name:      "test-project",
		RepoURL:   "https://github.com/test/repo",
		LocalPath: "/tmp/test",
		Status:    "active",
	}
	project.ID = "proj-001"

	// Run backup.
	ctx := t.Context()
	out, err := svc.Execute(ctx, services.BackupInput{
		Trigger:    "manual",
		EntityType: "project",
		EntityID:   project.ID,
		EntityData: project,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.FilePath)
	assert.FileExists(t, out.FilePath)

	// Read the file back.
	data, err := os.ReadFile(out.FilePath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test-project")
	assert.Contains(t, string(data), "https://github.com/test/repo")

	// Recover all.
	recovered, err := svc.Recover(ctx)
	require.NoError(t, err)
	assert.Len(t, recovered, 1)

	// Verify recovered data.
	rec, ok := recovered[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "test-project", rec["name"])
	assert.Equal(t, "active", rec["status"])
}

// TestE2E_ReviewOutputParsing tests review structured output parsing from mock scripts.
func TestE2E_ReviewOutputParsing(t *testing.T) {
	t.Run("pass", func(t *testing.T) {
		out := runMockScript(t, "mock_claude_reviewer_pass.sh")
		var result claude.Result
		require.NoError(t, json.Unmarshal(out, &result))

		var review struct {
			Verdict string `json:"verdict"`
		}
		require.NoError(t, json.Unmarshal(result.StructuredOutput, &review))
		assert.Equal(t, "pass", review.Verdict)
	})

	t.Run("fail", func(t *testing.T) {
		out := runMockScript(t, "mock_claude_reviewer_fail.sh")
		var result claude.Result
		require.NoError(t, json.Unmarshal(out, &result))

		var review struct {
			Verdict      string `json:"verdict"`
			AntiPatterns []string `json:"anti_patterns"`
		}
		require.NoError(t, json.Unmarshal(result.StructuredOutput, &review))
		assert.Equal(t, "fail", review.Verdict)
		assert.NotEmpty(t, review.AntiPatterns)
	})
}

// TestE2E_QuestionDetection verifies question detection in agent output.
func TestE2E_QuestionDetection(t *testing.T) {
	out := runMockScript(t, "mock_claude_question.sh")
	var result claude.Result
	require.NoError(t, json.Unmarshal(out, &result))

	assert.Contains(t, result.Text, "QUESTION:")
	assert.Contains(t, result.Text, "soft deletes or hard deletes")
}

// TODO: These tests require a running PostgreSQL database.
// Run with: DATABASE_URL=postgres://... go test -tags=e2e ./...
// TestE2E_FullWorkflow — PRD -> plan -> execute -> review -> complete
// TestE2E_RetryOnFailure — task fails -> retries -> succeeds
// TestE2E_CascadeFailure — task fails max retries -> deps cascade fail
// TestE2E_ModelEscalation — retry with escalated model
