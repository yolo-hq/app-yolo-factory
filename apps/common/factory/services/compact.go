package services

import (
	"context"
	"fmt"
	"strings"
)

// compactPlanOutput asks a small/cheap model to distill a plan-phase
// transcript down to the parts the implement phase actually needs:
// file list, decisions, test names. Reasoning and tool traces are dropped.
//
// On any failure, returns a truncated copy of the original output so the
// implement phase still has something to work with.
func (s *OrchestratorService) compactPlanOutput(ctx context.Context, workDir, planOutput string) string {
	planOutput = strings.TrimSpace(planOutput)
	if planOutput == "" {
		return ""
	}

	// Skip the round-trip when the plan is already small.
	if len(planOutput) < 2000 {
		return planOutput
	}

	prompt := "Summarize the plan below for an implementation agent.\n\n" +
		"Keep ONLY:\n" +
		"- Files to create or modify (with one-line purpose each)\n" +
		"- Concrete decisions (data shapes, function signatures, naming)\n" +
		"- Test names / acceptance criteria\n" +
		"- Open questions, if any\n\n" +
		"Drop reasoning, tool call traces, and exploration notes. " +
		"Be terse. Use bullet points.\n\n" +
		"--- PLAN ---\n" + planOutput

	cfg := phaseConfig("compact")
	cfg.CWD = workDir

	res, err := s.Claude.Run(ctx, cfg, prompt)
	if err != nil || res == nil || res.IsError || strings.TrimSpace(res.Text) == "" {
		// Fallback: keep first ~2k chars so impl still has context.
		const maxFallback = 2000
		if len(planOutput) > maxFallback {
			return planOutput[:maxFallback] + "\n... [truncated]"
		}
		return planOutput
	}
	return strings.TrimSpace(res.Text)
}

// withPlanSummary prepends a compacted plan summary to an implement prompt.
func withPlanSummary(implPrompt, planSummary string) string {
	if planSummary == "" {
		return implPrompt
	}
	return fmt.Sprintf("## Plan summary\n\n%s\n\n---\n\n%s", planSummary, implPrompt)
}
