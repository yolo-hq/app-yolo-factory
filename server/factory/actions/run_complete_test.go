package actions

import (
	"testing"
)

func TestParseDeps(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"valid array", `["task-1","task-2"]`, []string{"task-1", "task-2"}},
		{"single item", `["task-1"]`, []string{"task-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDeps(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseDeps(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseDeps(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseDeps_Empty(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"null", "null"},
		{"empty array", "[]"},
		{"invalid json", "{not json}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDeps(tt.input)
			if got != nil {
				t.Errorf("parseDeps(%q) = %v, want nil", tt.input, got)
			}
		})
	}
}

func TestContainsDep(t *testing.T) {
	deps := []string{"a", "b", "c"}
	if !containsDep(deps, "b") {
		t.Error("containsDep should find 'b'")
	}
	if containsDep(deps, "d") {
		t.Error("containsDep should not find 'd'")
	}
	if containsDep(nil, "a") {
		t.Error("containsDep(nil) should return false")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate short string: got %q", got)
	}
	if got := truncate("hello world", 5); got != "hello" {
		t.Errorf("truncate long string: got %q", got)
	}
	if got := truncate("", 5); got != "" {
		t.Errorf("truncate empty: got %q", got)
	}
}

// Integration test outline for the full CompleteRunAction flow.
// Actual DB-backed tests tracked in #34. The state machine logic:
//
// completed flow:
//   1. Run updated with metrics + completed_at
//   2. Task → "done", cost accumulated, summary stored
//   3. Blocked dependents with all deps met → "queued"
//   4. PRD counters recalculated from task statuses
//   5. PRD → "completed" when all tasks terminal
//   6. Next queued task dispatched via ExecuteWorkflowJob
//
// failed flow:
//   1. Run updated with metrics + error
//   2. If run_count < max_retries: task → "queued", job re-enqueued
//   3. If retries exhausted: task → "failed"
//   4. Cascade: downstream dependents → "failed" (recursive)
//   5. PRD counters updated, PRD → "failed" if all terminal with any failures
