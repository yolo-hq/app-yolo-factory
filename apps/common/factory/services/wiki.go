package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"
	yolostrings "github.com/yolo-hq/yolo/core/strings"
)

const (
	wikiMaxLines = 200
	wikiDir      = ".factory"
	wikiFile     = "wiki.md"
)

// wikiTemplate is the initial wiki content for new projects.
const wikiTemplate = `# Project Wiki (auto-maintained by Factory)
<!-- Last updated: %s, task: %s -->

## Architecture

## Conventions

## Gotchas

## Dependencies
`

// WikiService maintains a per-project knowledge base at {project}/.factory/wiki.md.
// It uses a Haiku agent to extract observations from task output and keep the wiki compact.
type WikiService struct {
	service.Base
	Claude *claude.Client
}

// WikiInput holds the data needed to update the wiki.
type WikiInput struct {
	ProjectPath  string
	TaskID       string
	TaskTitle    string
	TaskOutput   string // agent output, will be truncated
	ReviewOutput string // review output, will be truncated
}

// Update reads the existing wiki, calls Haiku to extract new observations,
// appends them, compacts if over 200 lines, and writes back.
// Non-fatal: logs errors but does not fail the caller.
func (s *WikiService) Update(ctx context.Context, in WikiInput) error {
	wikiPath := filepath.Join(in.ProjectPath, wikiDir, wikiFile)

	// Ensure .factory/ dir exists.
	if err := os.MkdirAll(filepath.Join(in.ProjectPath, wikiDir), 0o755); err != nil {
		return fmt.Errorf("mkdir .factory: %w", err)
	}

	// Read existing wiki (or initialize).
	existing := readWiki(wikiPath)
	if existing == "" {
		existing = fmt.Sprintf(wikiTemplate, time.Now().Format("2006-01-02"), in.TaskID)
		if err := os.WriteFile(wikiPath, []byte(existing), 0o644); err != nil {
			return fmt.Errorf("init wiki: %w", err)
		}
	}

	// Build Haiku prompt.
	taskSnippet := yolostrings.Truncate(in.TaskOutput, 1500)
	reviewSnippet := yolostrings.Truncate(in.ReviewOutput, 500)

	prompt := fmt.Sprintf(`You are maintaining a project wiki for an AI coding agent.

## Completed task
Title: %s

## Agent output (truncated)
%s

## Review output (truncated)
%s

## Current wiki
%s

## Instructions
Update the wiki above with new observations from the completed task.
Rules:
- Only add information that is genuinely new and useful for future tasks
- Keep each section focused: Architecture (structure/patterns), Conventions (naming/style), Gotchas (traps/pitfalls), Dependencies (libs/tools)
- Be terse: one line per fact
- Update the <!-- Last updated --> comment to: <!-- Last updated: %s, task: %s -->
- Return the FULL updated wiki (all sections, header included)
- Maximum %d lines total — compact aggressively if near limit`,
		in.TaskTitle,
		taskSnippet,
		reviewSnippet,
		existing,
		time.Now().Format("2006-01-02"),
		in.TaskID,
		wikiMaxLines,
	)

	cfg := phaseConfig("compact")
	cfg.CWD = in.ProjectPath

	res, err := s.Claude.Run(ctx, cfg, prompt)
	if err != nil || res == nil || res.IsError || strings.TrimSpace(res.Text) == "" {
		// Non-fatal: wiki stays as-is.
		return nil
	}

	updated := strings.TrimSpace(res.Text)

	// Enforce max lines via compaction if agent didn't comply.
	if lineCount(updated) > wikiMaxLines {
		updated = compactWiki(updated, wikiMaxLines)
	}

	if err := os.WriteFile(wikiPath, []byte(updated+"\n"), 0o644); err != nil {
		return fmt.Errorf("write wiki: %w", err)
	}
	return nil
}

// readWiki reads the wiki file. Returns empty string if not found.
func readWiki(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// readProjectWiki reads .factory/wiki.md from a project path. Returns empty string on error.
func readProjectWiki(projectPath string) string {
	if projectPath == "" {
		return ""
	}
	return readWiki(filepath.Join(projectPath, wikiDir, wikiFile))
}

// lineCount counts newline-separated lines.
func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return len(strings.Split(s, "\n"))
}

// compactWiki keeps the header + section headers and truncates body lines
// to fit within maxLines. Preserves section structure.
func compactWiki(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}

	// Keep first 2 lines (title + comment) plus section headers always.
	var keep []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "<!--") {
			keep = append(keep, line)
		} else if trimmed != "" && len(keep) < maxLines {
			keep = append(keep, line)
		}
	}

	// Trim to maxLines.
	if len(keep) > maxLines {
		keep = keep[:maxLines]
	}
	return strings.Join(keep, "\n")
}

func (s *WikiService) Description() string {
	return "Maintain per-project knowledge wiki from task observations"
}
