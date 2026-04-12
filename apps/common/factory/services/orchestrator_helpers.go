package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/oklog/ulid/v2"
	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	yolostrings "github.com/yolo-hq/yolo/core/strings"
)

// workingDir returns the directory to run agents in.
func workingDir(project entities.Project, taskID string) string {
	if project.UseWorktrees {
		return filepath.Join(project.LocalPath, ".worktrees", "task-"+taskID)
	}
	return project.LocalPath
}

// determineModel returns the model to use:
// task override > escalation (after retries) > project default > "sonnet".
func determineModel(task entities.Task, project entities.Project) string {
	if task.Model != "" {
		return task.Model
	}
	if task.RunCount >= project.EscalationAfterRetries && project.EscalationModel != "" {
		return project.EscalationModel
	}
	if project.DefaultModel != "" {
		return project.DefaultModel
	}
	return "sonnet"
}

// checkBudget verifies the project has not exceeded its monthly budget.
func checkBudget(project entities.Project) error {
	if project.BudgetMonthlyUSD > 0 && project.SpentThisMonthUSD >= project.BudgetMonthlyUSD {
		return fmt.Errorf("monthly budget exceeded: spent $%.2f of $%.2f limit",
			project.SpentThisMonthUSD, project.BudgetMonthlyUSD)
	}
	return nil
}

// detectQuestion scans agent output for "QUESTION:" prefix and extracts the question.
func detectQuestion(resultText string, task entities.Task, runID string) *entities.Question {
	upper := strings.ToUpper(resultText)
	idx := strings.Index(upper, "QUESTION:")
	if idx == -1 {
		return nil
	}
	questionText := strings.TrimSpace(resultText[idx+9:])
	if nl := strings.Index(questionText, "\n"); nl > 0 {
		questionText = questionText[:nl]
	}
	if questionText == "" {
		return nil
	}

	q := &entities.Question{
		TaskID:     task.ID,
		RunID:      runID,
		Body:       questionText,
		Context:    "Detected during implementation step",
		Confidence: string(enums.QuestionConfidenceMedium),
		Status:     string(enums.QuestionStatusOpen),
	}
	q.ID = ulid.Make().String()
	return q
}

// parseTestCommands parses the project's test_commands JSON array.
func parseTestCommands(raw string) []string {
	if raw == "" || raw == "[]" {
		return []string{"go build ./...", "go test ./..."}
	}
	var cmds []string
	if err := json.Unmarshal([]byte(raw), &cmds); err != nil {
		return []string{"go build ./...", "go test ./..."}
	}
	if len(cmds) == 0 {
		return []string{"go build ./...", "go test ./..."}
	}
	return cmds
}

// shellMetaChars are characters that indicate shell injection attempts.
var shellMetaChars = []string{"|", ";", "&", "$", "`", "(", ")", "{", "}", "<", ">", "!", "~"}

// validateCommand checks that a command string has no shell metacharacters.
func validateCommand(command string) error {
	for _, ch := range shellMetaChars {
		if strings.Contains(command, ch) {
			return fmt.Errorf("command contains shell metacharacter %q: %s", ch, command)
		}
	}
	return nil
}

// runShellCommand executes a command by splitting on whitespace (no shell).
func runShellCommand(ctx context.Context, dir string, command string) (string, error) {
	if err := validateCommand(command); err != nil {
		return "", err
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String() + stderr.String(), err
	}
	return stdout.String(), nil
}

// auditOutput is the structured output from the audit agent.
type auditOutput struct {
	Passed     bool     `json:"passed"`
	Violations []string `json:"violations"`
	Warnings   []string `json:"warnings"`
}

// parseAuditOutput parses audit structured output, returns (failed, errorMsg).
func parseAuditOutput(raw json.RawMessage) (bool, string) {
	if len(raw) == 0 {
		return false, ""
	}
	var out auditOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return true, fmt.Sprintf("failed to parse audit output: %s", err)
	}
	if !out.Passed {
		return true, fmt.Sprintf("audit failed: %s", strings.Join(out.Violations, "; "))
	}
	return false, ""
}

// reviewOutput is the structured output from the review agent.
type reviewOutput struct {
	Verdict         string          `json:"verdict"`
	CriteriaResults json.RawMessage `json:"criteria_results"`
	AntiPatterns    []string        `json:"anti_patterns"`
	Reasons         []string        `json:"reasons"`
	Suggestions     []string        `json:"suggestions"`
}

// parseReviewOutput parses review structured output into a Review entity.
func parseReviewOutput(raw json.RawMessage, runID, taskID string, result *claude.Result) (*entities.Review, bool, string) {
	if len(raw) == 0 {
		return nil, false, ""
	}
	var out reviewOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, true, fmt.Sprintf("failed to parse review output: %s", err)
	}

	review := &entities.Review{
		RunID:           runID,
		TaskID:          taskID,
		SessionID:       result.SessionID,
		Model:           "sonnet",
		Verdict:         out.Verdict,
		Reasons:         helpers.ToJSON(out.Reasons),
		AntiPatterns:    helpers.ToJSON(out.AntiPatterns),
		CriteriaResults: string(out.CriteriaResults),
		Suggestions:     helpers.ToJSON(out.Suggestions),
		CostUSD:         result.CostUSD,
	}
	review.ID = ulid.Make().String()

	if out.Verdict == string(enums.ReviewVerdictFail) {
		return review, true, fmt.Sprintf("review failed: %s", strings.Join(out.Reasons, "; "))
	}
	return review, false, ""
}

// toJSON is a package-local alias used within the orchestrator files.
func toJSON(v interface{}) string {
	return helpers.ToJSON(v)
}

// truncate500 truncates a string to 500 characters.
func truncate500(s string) string {
	return yolostrings.Truncate(s, 500)
}
