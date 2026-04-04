package services

import (
	"context"
	"strings"
	"testing"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

func TestContext_PlanTasks(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase: "plan_tasks",
		Project: entities.Project{
			Name: "yolo-core",
		},
		PRD: entities.PRD{
			Title:              "Add retry policies",
			Body:               "Implement configurable retry policies for all job types.",
			AcceptanceCriteria: `[{"id":"ac-1","description":"RetryPolicy interface exists","verification":"go build"}]`,
			DesignDecisions:    `["Exponential backoff default","Max 5 retries"]`,
		},
		Task: entities.Task{
			Branch: "main",
		},
		CLAUDEMDContent: "# Rules\nIntegration tests only",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, out.Prompt, "yolo-core")
	assertContains(t, out.Prompt, "Add retry policies")
	assertContains(t, out.Prompt, "configurable retry policies")
	assertContains(t, out.Prompt, "[ac-1] RetryPolicy interface exists")
	assertContains(t, out.Prompt, "Exponential backoff default")
	assertContains(t, out.Prompt, "Integration tests only")
	assertContains(t, out.SystemPrompt, "architect")
}

func TestContext_Implement(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase: "implement",
		Task: entities.Task{
			Title:              "Add RetryPolicy interface",
			Spec:               "Create RetryPolicy interface in core/retry/",
			AcceptanceCriteria: `[{"id":"tc-1","description":"Interface exists"}]`,
		},
		CompletedDeps: []DepSummary{
			{
				Title:        "Setup retry package",
				Summary:      "Created core/retry/ package structure",
				CommitHash:   "abc123",
				FilesChanged: []string{"core/retry/retry.go"},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, out.Prompt, "Add RetryPolicy interface")
	assertContains(t, out.Prompt, "core/retry/")
	assertContains(t, out.Prompt, "[tc-1] Interface exists")
	assertContains(t, out.Prompt, "Setup retry package")
	assertContains(t, out.Prompt, "abc123")
	assertContains(t, out.Prompt, "core/retry/retry.go")
}

func TestContext_ImplementRetry(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase: "implement",
		Task: entities.Task{
			Title:              "Fix broken build",
			Spec:               "Fix compilation error in retry package",
			AcceptanceCriteria: "go build passes",
		},
		IsRetry:       true,
		RetryError:    "undefined: RetryPolicy",
		ReviewReasons: "Missing import statement",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, out.Prompt, "Previous Attempt Failed")
	assertContains(t, out.Prompt, "undefined: RetryPolicy")
	assertContains(t, out.Prompt, "Missing import statement")
}

func TestContext_ReviewTask(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase: "review_task",
		Task: entities.Task{
			Title:              "Add RetryPolicy",
			Spec:               "Create interface",
			AcceptanceCriteria: `[{"id":"tc-1","description":"Interface exists"}]`,
		},
		GitDiff: "+type RetryPolicy interface {\n+  ShouldRetry(attempt int) bool\n+}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, out.Prompt, "Add RetryPolicy")
	assertContains(t, out.Prompt, "[tc-1] Interface exists")
	assertContains(t, out.Prompt, "+type RetryPolicy interface")
	assertContains(t, out.Prompt, "Anti-Pattern Checklist")
}

func TestContext_Audit(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase:           "audit",
		ChangedFiles:    "core/retry/policy.go\ncore/retry/policy_test.go",
		CLAUDEMDContent: "# Rules\nIntegration tests only\nNo mocks",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, out.Prompt, "core/retry/policy.go")
	assertContains(t, out.Prompt, "Integration tests only")
	assertContains(t, out.Prompt, "No mocks")
}

func TestContext_ReviewPRD(t *testing.T) {
	svc := &ContextService{}
	out, err := svc.Execute(context.Background(), ContextInput{
		Phase: "review_prd",
		PRD: entities.PRD{
			Title:              "Add retry policies",
			Body:               "Full retry system for jobs.",
			AcceptanceCriteria: `[{"id":"ac-1","description":"Retry works","verification":"run tests"}]`,
		},
		TaskSummaries: "### 1. Add RetryPolicy interface\nSummary: Created interface\nCommit: abc123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertContains(t, out.Prompt, "Add retry policies")
	assertContains(t, out.Prompt, "Full retry system")
	assertContains(t, out.Prompt, "[ac-1] Retry works")
	assertContains(t, out.Prompt, "Add RetryPolicy interface")
}

func TestContext_UnknownPhase(t *testing.T) {
	svc := &ContextService{}
	_, err := svc.Execute(context.Background(), ContextInput{
		Phase: "unknown",
	})
	if err == nil {
		t.Fatal("expected error for unknown phase")
	}
	assertContains(t, err.Error(), "unknown phase")
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, s)
	}
}
