package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
)

// QuestionResolverService attempts to answer agent questions automatically.
type QuestionResolverService struct {
	service.Base
	Claude *claude.Client
}

// QuestionResolverInput holds the question and its context.
type QuestionResolverInput struct {
	Question entities.Question
	Task     entities.Task
	Project  entities.Project
}

// QuestionResolverOutput holds the resolution result.
type QuestionResolverOutput struct {
	Answer     string
	AnsweredBy string // "auto" | "planner" | "" (needs human)
	Resolved   bool
	SessionID  string
}

// Execute tries to resolve a question: first via CLAUDE.md, then via planner agent.
func (s *QuestionResolverService) Execute(ctx context.Context, in QuestionResolverInput) (QuestionResolverOutput, error) {
	// 1. Auto-resolve: search CLAUDE.md for matching keywords.
	claudeMD := readCLAUDEMD(in.Project.LocalPath)
	if answer := autoResolve(claudeMD, in.Question.Body); answer != "" {
		return QuestionResolverOutput{
			Answer:     answer,
			AnsweredBy: "auto",
			Resolved:   true,
		}, nil
	}

	// 2. Ask planner agent.
	if s.Claude != nil {
		result, err := s.askPlanner(ctx, in, claudeMD)
		if err != nil {
			return QuestionResolverOutput{}, fmt.Errorf("ask planner: %w", err)
		}
		if result.Resolved {
			return result, nil
		}
	}

	// 3. Needs human — emit event.
	events.QuestionNeedsHuman.Emit(ctx, in.Question.ID)
	return QuestionResolverOutput{
		Resolved: false,
	}, nil
}

// autoResolve searches CLAUDE.md content for lines matching question keywords.
// Returns matching lines if a clear match is found.
func autoResolve(claudeMD, questionBody string) string {
	if claudeMD == "" || questionBody == "" {
		return ""
	}

	// Extract keywords from question (words >= 4 chars, lowered).
	keywords := extractKeywords(questionBody)
	if len(keywords) == 0 {
		return ""
	}

	// Search CLAUDE.md lines for matching keywords.
	lines := strings.Split(claudeMD, "\n")
	var matches []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		matchCount := 0
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				matchCount++
			}
		}
		// Require at least 2 keyword matches or 1 match if few keywords.
		threshold := 2
		if len(keywords) <= 2 {
			threshold = 1
		}
		if matchCount >= threshold {
			matches = append(matches, strings.TrimSpace(line))
		}
	}

	if len(matches) == 0 {
		return ""
	}

	return "From CLAUDE.md:\n" + strings.Join(matches, "\n")
}

// extractKeywords pulls significant words from text (len >= 4, lowercased).
func extractKeywords(text string) []string {
	words := strings.Fields(text)
	seen := make(map[string]bool)
	var keywords []string

	// Common stop words to skip.
	stop := map[string]bool{
		"what": true, "when": true, "where": true, "which": true,
		"that": true, "this": true, "with": true, "from": true,
		"have": true, "does": true, "should": true, "could": true,
		"would": true, "about": true, "there": true, "their": true,
		"been": true, "will": true, "into": true, "also": true,
	}

	for _, w := range words {
		// Strip punctuation.
		w = strings.Trim(w, ".,;:!?\"'()[]{}/-")
		w = strings.ToLower(w)
		if len(w) < 4 || stop[w] || seen[w] {
			continue
		}
		seen[w] = true
		keywords = append(keywords, w)
	}
	return keywords
}

// askPlanner spawns an Opus agent to answer the question.
func (s *QuestionResolverService) askPlanner(ctx context.Context, in QuestionResolverInput, claudeMD string) (QuestionResolverOutput, error) {
	prompt := fmt.Sprintf(`Answer this question from an agent working on task "%s" in project "%s".

## Question
%s

## Question Context
%s

## Project CLAUDE.md
%s

If you can answer confidently, provide a clear answer.
If you are UNSURE, respond with exactly "UNSURE" at the start of your response.`,
		in.Task.Title, in.Project.Name,
		in.Question.Body, in.Question.Context, claudeMD)

	result, err := s.Claude.Run(ctx, claude.Config{
		Model:          "opus",
		AllowedTools:   []string{"Read", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      0.50,
		PermissionMode: "auto",
		Effort:         "high",
		CWD:            in.Project.LocalPath,
		SessionName:    fmt.Sprintf("factory:question-%s:resolve", in.Question.ID),
		Timeout:        5 * time.Minute,
	}, prompt)
	if err != nil {
		return QuestionResolverOutput{}, err
	}

	if result.IsError {
		return QuestionResolverOutput{}, fmt.Errorf("claude error: %s", result.Text)
	}

	// Check if planner is unsure.
	if strings.HasPrefix(strings.TrimSpace(result.Text), "UNSURE") {
		return QuestionResolverOutput{Resolved: false}, nil
	}

	return QuestionResolverOutput{
		Answer:     result.Text,
		AnsweredBy: "planner",
		Resolved:   true,
		SessionID:  result.SessionID,
	}, nil
}

func (s *QuestionResolverService) Description() string { return "Attempt auto-resolution of open questions" }
