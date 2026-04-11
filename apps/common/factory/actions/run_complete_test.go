package actions

import (
	"testing"

	yolostrings "github.com/yolo-hq/yolo/core/strings"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
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
			got := helpers.ParseDeps(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseDeps(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseDeps(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
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
			got := helpers.ParseDeps(tt.input)
			if got != nil {
				t.Errorf("ParseDeps(%q) = %v, want nil", tt.input, got)
			}
		})
	}
}

func TestContainsDep(t *testing.T) {
	depsJSON := `["a","b","c"]`
	if !helpers.ContainsDep(depsJSON, "b") {
		t.Error("ContainsDep should find 'b'")
	}
	if helpers.ContainsDep(depsJSON, "d") {
		t.Error("ContainsDep should not find 'd'")
	}
	if helpers.ContainsDep("[]", "a") {
		t.Error("ContainsDep([]) should return false")
	}
}

func TestTruncate(t *testing.T) {
	if got := yolostrings.Truncate("hello", 10); got != "hello" {
		t.Errorf("Truncate short string: got %q", got)
	}
	if got := yolostrings.Truncate("hello world", 5); got != "hello..." {
		t.Errorf("Truncate long string: got %q", got)
	}
	if got := yolostrings.Truncate("", 5); got != "" {
		t.Errorf("Truncate empty: got %q", got)
	}
}

// Integration test outline for the full CompleteRunAction flow.
// Actual DB-backed tests tracked in #34. The state machine logic:
//
// completed flow:
//  1. Run updated with metrics + completed_at
//  2. Task → "done", cost accumulated, summary stored
//  3. Blocked dependents with all deps met → "queued"
//  4. PRD counters recalculated from task statuses
//  5. PRD → "completed" when all tasks terminal
//  6. Next queued task dispatched via ExecuteWorkflowJob
//
// failed flow:
//  1. Run updated with metrics + error
//  2. If run_count < max_retries: task → "queued", job re-enqueued
//  3. If retries exhausted: task → "failed"
//  4. Cascade: downstream dependents → "failed" (recursive)
//  5. PRD counters updated, PRD → "failed" if all terminal with any failures
