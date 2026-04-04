package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yolo-hq/yolo/core/pkg/claude"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

var mockResult = claude.Result{
	SessionID: "test-session",
	CostUSD:   0.05,
	Usage:     claude.Usage{InputTokens: 100, OutputTokens: 50},
}

func TestDetermineModel_TaskOverride(t *testing.T) {
	task := entities.Task{Model: "opus"}
	project := entities.Project{DefaultModel: "sonnet"}

	got := determineModel(task, project)
	assert.Equal(t, "opus", got)
}

func TestDetermineModel_ProjectDefault(t *testing.T) {
	task := entities.Task{}
	project := entities.Project{DefaultModel: "haiku"}

	got := determineModel(task, project)
	assert.Equal(t, "haiku", got)
}

func TestDetermineModel_Fallback(t *testing.T) {
	task := entities.Task{}
	project := entities.Project{}

	got := determineModel(task, project)
	assert.Equal(t, "sonnet", got)
}

func TestWorkingDir_Direct(t *testing.T) {
	project := entities.Project{
		LocalPath:    "/home/user/project",
		UseWorktrees: false,
	}

	got := workingDir(project, "abc123")
	assert.Equal(t, "/home/user/project", got)
}

func TestWorkingDir_Worktree(t *testing.T) {
	project := entities.Project{
		LocalPath:    "/home/user/project",
		UseWorktrees: true,
	}

	got := workingDir(project, "abc123")
	assert.Equal(t, "/home/user/project/.worktrees/task-abc123", got)
}

func TestParseTestCommands_ValidJSON(t *testing.T) {
	raw := `["make build", "make test", "make lint"]`
	got := parseTestCommands(raw)
	assert.Equal(t, []string{"make build", "make test", "make lint"}, got)
}

func TestParseTestCommands_Empty(t *testing.T) {
	got := parseTestCommands("")
	assert.Equal(t, []string{"go build ./...", "go test ./..."}, got)

	got = parseTestCommands("[]")
	assert.Equal(t, []string{"go build ./...", "go test ./..."}, got)
}

func TestParseTestCommands_InvalidJSON(t *testing.T) {
	got := parseTestCommands("not json")
	assert.Equal(t, []string{"go build ./...", "go test ./..."}, got)
}

func TestTruncateSummary_Short(t *testing.T) {
	got := Truncate("short text", 500)
	assert.Equal(t, "short text", got)
}

func TestTruncateSummary_Long(t *testing.T) {
	long := ""
	for i := 0; i < 100; i++ {
		long += "abcdefghij" // 1000 chars
	}
	got := Truncate(long, 500)
	assert.Len(t, got, 503) // 500 + "..."
	assert.Equal(t, long[:500]+"...", got)
}

func TestTruncateSummary_ExactLength(t *testing.T) {
	text := "12345"
	got := Truncate(text, 5)
	assert.Equal(t, "12345", got)
}

func TestParseAuditOutput_Passed(t *testing.T) {
	raw := []byte(`{"passed": true, "violations": [], "warnings": ["minor thing"]}`)
	failed, errMsg := parseAuditOutput(raw)
	assert.False(t, failed)
	assert.Empty(t, errMsg)
}

func TestParseAuditOutput_Failed(t *testing.T) {
	raw := []byte(`{"passed": false, "violations": ["missing test", "wrong pattern"], "warnings": []}`)
	failed, errMsg := parseAuditOutput(raw)
	assert.True(t, failed)
	assert.Contains(t, errMsg, "missing test")
	assert.Contains(t, errMsg, "wrong pattern")
}

func TestParseAuditOutput_Empty(t *testing.T) {
	failed, errMsg := parseAuditOutput(nil)
	assert.False(t, failed)
	assert.Empty(t, errMsg)
}

func TestParseReviewOutput_Pass(t *testing.T) {
	raw := []byte(`{
		"verdict": "pass",
		"criteria_results": [{"criteria_id": "AC1", "passed": true, "comment": "ok"}],
		"anti_patterns": [],
		"reasons": [],
		"suggestions": ["could improve naming"]
	}`)

	review, failed, errMsg := parseReviewOutput(raw, "run-1", "task-1", &mockResult)
	assert.NotNil(t, review)
	assert.False(t, failed)
	assert.Empty(t, errMsg)
	assert.Equal(t, "pass", review.Verdict)
	assert.Equal(t, "run-1", review.RunID)
	assert.Equal(t, "task-1", review.TaskID)
	assert.NotEmpty(t, review.ID)
}

func TestParseReviewOutput_Fail(t *testing.T) {
	raw := []byte(`{
		"verdict": "fail",
		"criteria_results": [{"criteria_id": "AC1", "passed": false, "comment": "broken"}],
		"anti_patterns": ["god function"],
		"reasons": ["tests missing", "wrong pattern"],
		"suggestions": []
	}`)

	review, failed, errMsg := parseReviewOutput(raw, "run-1", "task-1", &mockResult)
	assert.NotNil(t, review)
	assert.True(t, failed)
	assert.Contains(t, errMsg, "tests missing")
	assert.Contains(t, errMsg, "wrong pattern")
	assert.Equal(t, "fail", review.Verdict)
}

func TestParseReviewOutput_Empty(t *testing.T) {
	review, failed, errMsg := parseReviewOutput(nil, "run-1", "task-1", &mockResult)
	assert.Nil(t, review)
	assert.False(t, failed)
	assert.Empty(t, errMsg)
}
