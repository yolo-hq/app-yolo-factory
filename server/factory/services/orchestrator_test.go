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

// --- determineModel escalation tests ---

func TestDetermineModel_Default(t *testing.T) {
	task := entities.Task{}
	project := entities.Project{}
	got := determineModel(task, project)
	assert.Equal(t, "sonnet", got)
}

func TestDetermineModel_Escalation(t *testing.T) {
	task := entities.Task{RunCount: 3}
	project := entities.Project{
		DefaultModel:           "sonnet",
		EscalationModel:        "opus",
		EscalationAfterRetries: 2,
	}
	got := determineModel(task, project)
	assert.Equal(t, "opus", got)
}

func TestDetermineModel_NoEscalationBeforeThreshold(t *testing.T) {
	task := entities.Task{RunCount: 1}
	project := entities.Project{
		DefaultModel:           "sonnet",
		EscalationModel:        "opus",
		EscalationAfterRetries: 2,
	}
	got := determineModel(task, project)
	assert.Equal(t, "sonnet", got)
}

// --- checkBudget tests ---

func TestCheckBudget_WithinLimit(t *testing.T) {
	project := entities.Project{BudgetMonthlyUSD: 200, SpentThisMonthUSD: 100}
	err := checkBudget(project)
	assert.NoError(t, err)
}

func TestCheckBudget_Exceeded(t *testing.T) {
	project := entities.Project{BudgetMonthlyUSD: 200, SpentThisMonthUSD: 200}
	err := checkBudget(project)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monthly budget exceeded")
}

func TestCheckBudget_NoLimit(t *testing.T) {
	project := entities.Project{BudgetMonthlyUSD: 0, SpentThisMonthUSD: 9999}
	err := checkBudget(project)
	assert.NoError(t, err)
}

// --- detectQuestion tests ---

func TestDetectQuestion_Found(t *testing.T) {
	text := "I implemented the feature.\nQUESTION: Should I use X or Y?\nDone."
	task := entities.Task{}
	task.ID = "task-1"

	q := detectQuestion(text, task, "run-1")
	assert.NotNil(t, q)
	assert.Equal(t, "Should I use X or Y?", q.Body)
	assert.Equal(t, "task-1", q.TaskID)
	assert.Equal(t, "run-1", q.RunID)
	assert.Equal(t, entities.QuestionOpen, q.Status)
	assert.Equal(t, entities.ConfidenceMedium, q.Confidence)
	assert.NotEmpty(t, q.ID)
}

func TestDetectQuestion_NotFound(t *testing.T) {
	text := "Everything looks good. No issues found."
	task := entities.Task{}
	task.ID = "task-1"

	q := detectQuestion(text, task, "run-1")
	assert.Nil(t, q)
}

func TestDetectQuestion_EmptyAfterPrefix(t *testing.T) {
	text := "Some output\nQUESTION:"
	task := entities.Task{}
	task.ID = "task-1"

	q := detectQuestion(text, task, "run-1")
	assert.Nil(t, q)
}
