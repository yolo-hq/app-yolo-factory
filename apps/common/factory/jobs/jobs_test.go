package jobs_test

import (
	"testing"

	"github.com/yolo-hq/yolo/core/jobs"

	jobsgen "github.com/yolo-hq/app-yolo-factory/.yolo/gen/adapters/apps/common/factory/jobs"
	userjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
)

// compile-time interface checks.
var (
	_ jobs.Handler = (*userjobs.ExecuteWorkflowJob)(nil)
	_ jobs.Handler = (*userjobs.PlanPRDJob)(nil)
	_ jobs.Handler = (*jobsgen.AdvisorJob)(nil)
	_ jobs.Handler = (*jobsgen.BackupSnapshotJob)(nil)
	_ jobs.Handler = (*jobsgen.CheckTimeoutsJob)(nil)
	_ jobs.Handler = (*jobsgen.ProcessAdvisorJob)(nil)
	_ jobs.Handler = (*jobsgen.ResetBudgetsJob)(nil)
	_ jobs.Handler = (*jobsgen.SentinelJob)(nil)
)

func TestJobs_NameAndDescription(t *testing.T) {
	cases := []struct {
		handler  jobs.Handler
		wantName string
	}{
		{&userjobs.ExecuteWorkflowJob{}, "factory.execute-workflow"},
		{&userjobs.PlanPRDJob{}, "factory.plan-prd"},
		{jobsgen.AdvisorJob{}, "factory.advisor"},
		{jobsgen.BackupSnapshotJob{}, "factory.backup-snapshot"},
		{jobsgen.CheckTimeoutsJob{}, "factory.check-timeouts"},
		{jobsgen.ProcessAdvisorJob{}, "factory.process-advisor"},
		{jobsgen.ResetBudgetsJob{}, "factory.reset-monthly-budgets"},
		{jobsgen.SentinelJob{}, "factory.sentinel"},
	}

	for _, tc := range cases {
		t.Run(tc.wantName, func(t *testing.T) {
			if got := tc.handler.Name(); got != tc.wantName {
				t.Errorf("Name() = %q, want %q", got, tc.wantName)
			}

			type describer interface{ Description() string }
			d, ok := tc.handler.(describer)
			if !ok {
				t.Fatalf("%T does not implement Description()", tc.handler)
			}
			if d.Description() == "" {
				t.Errorf("%T.Description() returned empty string", tc.handler)
			}

			cfg := tc.handler.Config()
			if cfg.Timeout <= 0 {
				t.Errorf("%T.Config().Timeout = %v, want > 0", tc.handler, cfg.Timeout)
			}
		})
	}
}
