package services

import (
	"time"

	"github.com/yolo-hq/yolo/core/pkg/claude"
)

// phaseDefaults holds the per-phase claude.Config baseline.
// Runtime fields (CWD, SessionName, Model overrides) are layered on
// top in the orchestrator. Tweak tools/budgets/turn caps here.
var phaseDefaults = map[string]claude.Config{
	"plan": {
		Model:          "opus",
		AllowedTools:   []string{"Read", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      1.0,
		PermissionMode: "auto",
		Effort:         "high",
		MaxTurns:       40,
		Timeout:        10 * time.Minute,
	},
	"implement": {
		AllowedTools:   []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      2.0,
		PermissionMode: "auto",
		Effort:         "high",
		MaxTurns:       80,
		Timeout:        15 * time.Minute,
	},
	"audit": {
		Model:          "sonnet",
		AllowedTools:   []string{"Read", "Bash", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      0.30,
		PermissionMode: "auto",
		Effort:         "medium",
		MaxTurns:       30,
		Timeout:        5 * time.Minute,
	},
	"review": {
		Model:          "sonnet",
		AllowedTools:   []string{"Read", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      0.50,
		PermissionMode: "auto",
		Effort:         "medium",
		MaxTurns:       30,
		Timeout:        5 * time.Minute,
	},
	"compact": {
		Model:          "haiku",
		Bare:           true,
		BudgetUSD:      0.05,
		PermissionMode: "auto",
		Effort:         "low",
		MaxTurns:       2,
		Timeout:        2 * time.Minute,
	},
}

// phaseConfig returns a copy of the baseline for a phase, ready to be
// customized with runtime fields like CWD and SessionName.
func phaseConfig(phase string) claude.Config {
	cfg := phaseDefaults[phase]
	// Copy slices so callers don't mutate the shared default.
	if len(cfg.AllowedTools) > 0 {
		tools := make([]string, len(cfg.AllowedTools))
		copy(tools, cfg.AllowedTools)
		cfg.AllowedTools = tools
	}
	return cfg
}
