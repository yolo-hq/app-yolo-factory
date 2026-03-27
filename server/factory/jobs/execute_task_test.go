package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCost_WithCostLine(t *testing.T) {
	output := `some output
{"total_cost_usd": 0.42, "other": "data"}
more output`
	assert.InDelta(t, 0.42, parseCost(output), 0.001)
}

func TestParseCost_NoCostLine(t *testing.T) {
	assert.Equal(t, 0.0, parseCost("no cost here"))
}

func TestParseCost_EmptyOutput(t *testing.T) {
	assert.Equal(t, 0.0, parseCost(""))
}

func TestComposeWorkerPrompt_AllSections(t *testing.T) {
	prompt := ComposeWorkerPrompt("# My CLAUDE.md", "Fix the bug", []string{"go test ./...", "go build ./..."})

	assert.Contains(t, prompt, "# FRAMEWORK CONTEXT")
	assert.Contains(t, prompt, "entity.BaseEntity")
	assert.Contains(t, prompt, "# REPO CONTEXT")
	assert.Contains(t, prompt, "# My CLAUDE.md")
	assert.Contains(t, prompt, "# TASK")
	assert.Contains(t, prompt, "Fix the bug")
	assert.Contains(t, prompt, "# FEEDBACK LOOPS")
	assert.Contains(t, prompt, "go test ./...")
	assert.Contains(t, prompt, "go build ./...")
	assert.Contains(t, prompt, "COMPLETE")
}

func TestComposeWorkerPrompt_NoClaudeMD(t *testing.T) {
	prompt := ComposeWorkerPrompt("", "Do the thing", nil)

	assert.Contains(t, prompt, "# FRAMEWORK CONTEXT")
	assert.NotContains(t, prompt, "# REPO CONTEXT")
	assert.Contains(t, prompt, "# TASK")
	assert.Contains(t, prompt, "Do the thing")
	assert.NotContains(t, prompt, "# FEEDBACK LOOPS")
}

func TestComposeWorkerPrompt_NoFeedbackLoops(t *testing.T) {
	prompt := ComposeWorkerPrompt("content", "task", nil)

	assert.Contains(t, prompt, "# REPO CONTEXT")
	assert.NotContains(t, prompt, "# FEEDBACK LOOPS")
}
