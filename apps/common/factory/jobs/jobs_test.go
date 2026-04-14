package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/yolo-hq/yolo/core/jobs"
)

// compile-time interface checks.
var (
	_ jobs.Handler = (*ExecuteWorkflowJob)(nil)
	_ jobs.Handler = (*PlanPRDJob)(nil)
	_ jobs.Handler = (*BackupSnapshotJob)(nil)
	_ jobs.Handler = (*CheckTimeoutsJob)(nil)
	_ jobs.Handler = (*ProcessAdvisorJob)(nil)
	_ jobs.Handler = (*ResetBudgetsJob)(nil)
	_ jobs.Handler = (*SentinelJob)(nil)
)

func TestJobs_NameAndDescription(t *testing.T) {
	cases := []struct {
		handler     jobs.Handler
		wantName    string
		wantDescLen int
	}{
		{&ExecuteWorkflowJob{}, "factory.execute-workflow", 1},
		{&PlanPRDJob{}, "factory.plan-prd", 1},
		{&BackupSnapshotJob{}, "factory.backup-snapshot", 1},
		{&CheckTimeoutsJob{}, "factory.check-timeouts", 1},
		{&ProcessAdvisorJob{}, "factory.process-advisor", 1},
		{&ResetBudgetsJob{}, "factory.reset-monthly-budgets", 1},
		{&SentinelJob{}, "factory.sentinel", 1},
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

func TestJobs_Config_Timeouts(t *testing.T) {
	// Execution jobs should have longer timeouts than utility jobs.
	execCfg := (&ExecuteWorkflowJob{}).Config()
	planCfg := (&PlanPRDJob{}).Config()
	checkCfg := (&CheckTimeoutsJob{}).Config()

	if execCfg.Timeout < 10*time.Minute {
		t.Errorf("ExecuteWorkflowJob timeout %v too short, want >= 10m", execCfg.Timeout)
	}
	if planCfg.Timeout < 5*time.Minute {
		t.Errorf("PlanPRDJob timeout %v too short, want >= 5m", planCfg.Timeout)
	}
	if checkCfg.Timeout > 2*time.Minute {
		t.Errorf("CheckTimeoutsJob timeout %v too long, want <= 2m", checkCfg.Timeout)
	}
}

func TestExecuteWorkflowJob_PayloadParsing(t *testing.T) {
	payload := map[string]string{"task_id": "task-abc"}
	data, _ := json.Marshal(payload)

	var p ExecuteWorkflowJob
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if p.TaskID != "task-abc" {
		t.Errorf("TaskID = %q, want %q", p.TaskID, "task-abc")
	}
}

func TestExecuteWorkflowJob_PayloadParsing_Invalid(t *testing.T) {
	var p ExecuteWorkflowJob
	if err := json.Unmarshal([]byte("{not json}"), &p); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestPlanPRDJob_PayloadParsing(t *testing.T) {
	payload := map[string]string{"prd_id": "prd-xyz"}
	data, _ := json.Marshal(payload)

	var p PlanPRDJob
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if p.PRDID != "prd-xyz" {
		t.Errorf("PRDID = %q, want %q", p.PRDID, "prd-xyz")
	}
}

func TestSentinelJob_PayloadParsing(t *testing.T) {
	payload := sentinelPayload{
		ProjectID: "proj-1",
		Watches:   []string{"build", "tests"},
	}
	data, _ := json.Marshal(payload)

	var p sentinelPayload
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if p.ProjectID != "proj-1" {
		t.Errorf("ProjectID = %q, want %q", p.ProjectID, "proj-1")
	}
	if len(p.Watches) != 2 {
		t.Errorf("len(Watches) = %d, want 2", len(p.Watches))
	}
}

func TestProcessAdvisorJob_PayloadParsing(t *testing.T) {
	payload := processAdvisorPayload{ProjectID: "proj-2"}
	data, _ := json.Marshal(payload)

	var p processAdvisorPayload
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if p.ProjectID != "proj-2" {
		t.Errorf("ProjectID = %q, want %q", p.ProjectID, "proj-2")
	}
}
