package services

import (
	"context"
	"strings"
	"testing"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

func TestContext_Sentinel(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase:   "sentinel",
		Project: entities.Project{Name: "test-project"},
		Watches: "- Build status\n- Test coverage",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if out.SystemPrompt == "" {
		t.Fatal("expected non-empty system prompt")
	}
	if !strings.Contains(out.Prompt, "test-project") {
		t.Fatal("expected prompt to contain project name")
	}
}

func TestContext_Advisor(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase:           "advisor",
		Project:         entities.Project{Name: "my-project"},
		AnalysisType:    "cost",
		AnalysisContext: "monthly review",
		RunHistory:      "3 runs, 2 passed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	if !strings.Contains(out.Prompt, "my-project") {
		t.Fatal("expected prompt to contain project name")
	}
	if !strings.Contains(out.Prompt, "cost") {
		t.Fatal("expected prompt to contain analysis type")
	}
}

func TestContext_UnknownPhase(t *testing.T) {
	svc := &ContextService{}
	_, err := svc.Execute(context.Background(), ContextInput{
		Phase: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for unknown phase")
	}
}

// assertContains is a shared test helper for checking substring presence.
func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected %q to contain %q", s, sub)
	}
}
