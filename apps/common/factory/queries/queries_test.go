package queries

import (
	"context"
	"testing"

	"github.com/yolo-hq/yolo/core/query"
)

// compile-time interface checks.
var (
	_ query.Query = (*CostQuery)(nil)
	_ query.Query = (*StatusQuery)(nil)
	_ query.Query = (*PrdDiffQuery)(nil)
)

func TestQueries_Description(t *testing.T) {
	cases := []struct {
		name string
		q    interface{ Description() string }
	}{
		{"CostQuery", &CostQuery{}},
		{"StatusQuery", &StatusQuery{}},
		{"PrdDiffQuery", &PrdDiffQuery{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.q.Description() == "" {
				t.Errorf("%T.Description() returned empty string", tc.q)
			}
		})
	}
}

func TestRunPRDGitDiff_InvalidRepo(t *testing.T) {
	// Passing a non-existent path should return an error.
	_, _, err := runPRDGitDiff(context.Background(), "/nonexistent/path", "abc123", "def456")
	if err == nil {
		t.Error("expected error for invalid repo path, got nil")
	}
}

func TestRunPRDGitDiff_SameCommit(t *testing.T) {
	// Using /tmp as a non-git dir should also fail gracefully.
	_, _, err := runPRDGitDiff(context.Background(), "/tmp", "deadbeef", "deadbeef")
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}
}
