package queries_test

import (
	"context"
	"testing"

	"github.com/yolo-hq/yolo/core/query"

	queriesgen "github.com/yolo-hq/app-yolo-factory/.yolo/gen/adapters/apps/common/factory/queries"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/queries"
)

// compile-time interface checks.
var (
	_ query.Query = (*queriesgen.CostQuery)(nil)
	_ query.Query = (*queriesgen.StatusQuery)(nil)
	_ query.Query = (*queriesgen.PRDDiffQuery)(nil)
)

func TestQueries_Description(t *testing.T) {
	cases := []struct {
		name string
		q    interface{ Description() string }
	}{
		{"CostQuery", queriesgen.CostQuery{}},
		{"StatusQuery", queriesgen.StatusQuery{}},
		{"PRDDiffQuery", queriesgen.PRDDiffQuery{}},
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
	_, _, err := queries.RunPRDGitDiffForTest(context.Background(), "/nonexistent/path", "abc123", "def456")
	if err == nil {
		t.Error("expected error for invalid repo path, got nil")
	}
}

func TestRunPRDGitDiff_SameCommit(t *testing.T) {
	_, _, err := queries.RunPRDGitDiffForTest(context.Background(), "/tmp", "deadbeef", "deadbeef")
	if err == nil {
		t.Error("expected error for non-git directory, got nil")
	}
}
