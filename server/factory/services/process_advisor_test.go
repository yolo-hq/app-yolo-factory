package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

func TestComputeMetrics_Basic(t *testing.T) {
	tasks := []entities.Task{
		{Status: entities.TaskDone, CostUSD: 1.50, RunCount: 1},
		{Status: entities.TaskDone, CostUSD: 2.00, RunCount: 2},
		{Status: entities.TaskFailed, CostUSD: 0.50, RunCount: 3},
	}

	m := ComputeMetrics(tasks, nil, nil, nil, nil)

	assert.Equal(t, 3, m.TotalTasks)
	assert.Equal(t, 2, m.CompletedTasks)
	assert.Equal(t, 1, m.FailedTasks)
	assert.InDelta(t, 0.6667, m.SuccessRate, 0.01)
	assert.InDelta(t, 2.0, m.AvgRetriesPerTask, 0.01)
	assert.InDelta(t, 4.0, m.TotalCostUSD, 0.01)
	assert.InDelta(t, 1.3333, m.AvgCostPerTask, 0.01)
}

func TestComputeMetrics_ModelStats(t *testing.T) {
	runs := []entities.Run{
		{Model: "sonnet", Status: entities.RunCompleted, CostUSD: 0.50},
		{Model: "sonnet", Status: entities.RunCompleted, CostUSD: 0.60},
		{Model: "sonnet", Status: entities.RunFailed, CostUSD: 0.40, Error: "build failed"},
		{Model: "opus", Status: entities.RunCompleted, CostUSD: 2.00},
	}

	m := ComputeMetrics(nil, runs, nil, nil, nil)

	require.Contains(t, m.ModelStats, "sonnet")
	require.Contains(t, m.ModelStats, "opus")

	sonnet := m.ModelStats["sonnet"]
	assert.Equal(t, 3, sonnet.Tasks)
	assert.Equal(t, 2, sonnet.Successes)
	assert.InDelta(t, 0.6667, sonnet.SuccessRate, 0.01)
	assert.InDelta(t, 0.50, sonnet.AvgCostUSD, 0.01)

	opus := m.ModelStats["opus"]
	assert.Equal(t, 1, opus.Tasks)
	assert.Equal(t, 1, opus.Successes)
	assert.InDelta(t, 1.0, opus.SuccessRate, 0.01)
	assert.InDelta(t, 2.0, opus.AvgCostUSD, 0.01)
}

func TestComputeMetrics_StepStats(t *testing.T) {
	steps := []entities.Step{
		{Phase: "implement", Status: entities.StepCompleted, CostUSD: 1.00, DurationMs: 5000},
		{Phase: "implement", Status: entities.StepFailed, CostUSD: 0.80, DurationMs: 3000},
		{Phase: "review", Status: entities.StepCompleted, CostUSD: 0.30, DurationMs: 2000},
	}

	m := ComputeMetrics(nil, nil, steps, nil, nil)

	require.Contains(t, m.StepStats, "implement")
	require.Contains(t, m.StepStats, "review")

	impl := m.StepStats["implement"]
	assert.InDelta(t, 0.90, impl.AvgCostUSD, 0.01)
	assert.Equal(t, 4000, impl.AvgDurationMS)
	assert.InDelta(t, 0.50, impl.FailRate, 0.01)

	rev := m.StepStats["review"]
	assert.InDelta(t, 0.30, rev.AvgCostUSD, 0.01)
	assert.InDelta(t, 0.0, rev.FailRate, 0.01)
}

func TestComputeMetrics_Empty(t *testing.T) {
	m := ComputeMetrics(nil, nil, nil, nil, nil)

	assert.Equal(t, 0, m.TotalTasks)
	assert.Equal(t, 0, m.CompletedTasks)
	assert.Equal(t, 0, m.FailedTasks)
	assert.Equal(t, 0.0, m.SuccessRate)
	assert.Equal(t, 0.0, m.AvgRetriesPerTask)
	assert.Equal(t, 0.0, m.TotalCostUSD)
	assert.Equal(t, 0.0, m.AvgCostPerTask)
	assert.Equal(t, 0.0, m.LintCatchRate)
	assert.Equal(t, 0.0, m.ReviewFailRate)
	assert.Empty(t, m.ModelStats)
	assert.Empty(t, m.StepStats)
	assert.Empty(t, m.ErrorBreakdown)
}

func TestComputeMetrics_LintAndReview(t *testing.T) {
	reviews := []entities.Review{
		{Verdict: entities.ReviewPass},
		{Verdict: entities.ReviewFail},
		{Verdict: entities.ReviewPass},
	}
	lintResults := []entities.LintResult{
		{Passed: true},
		{Passed: false},
		{Passed: false},
		{Passed: true},
	}

	m := ComputeMetrics(nil, nil, nil, reviews, lintResults)

	assert.InDelta(t, 0.3333, m.ReviewFailRate, 0.01)
	assert.InDelta(t, 0.50, m.LintCatchRate, 0.01)
}

func TestComputeMetrics_ErrorBreakdown(t *testing.T) {
	runs := []entities.Run{
		{Status: entities.RunFailed, Error: "build failed: exit code 1"},
		{Status: entities.RunFailed, Error: "test suite failed"},
		{Status: entities.RunFailed, Error: "build error in main.go"},
		{Status: entities.RunFailed, Error: "timeout exceeded"},
		{Status: entities.RunCompleted}, // not counted
	}

	m := ComputeMetrics(nil, runs, nil, nil, nil)

	assert.Equal(t, 2, m.ErrorBreakdown["build_error"])
	assert.Equal(t, 1, m.ErrorBreakdown["test_error"])
	assert.Equal(t, 1, m.ErrorBreakdown["timeout"])
}

func TestFormatMetrics(t *testing.T) {
	m := &ExecutionMetrics{
		TotalTasks:        30,
		CompletedTasks:    25,
		FailedTasks:       5,
		SuccessRate:       0.8333,
		AvgRetriesPerTask: 1.5,
		TotalCostUSD:      45.00,
		AvgCostPerTask:    1.50,
		ModelStats: map[string]ModelStat{
			"sonnet": {Tasks: 20, Successes: 18, AvgCostUSD: 0.80, SuccessRate: 0.90},
		},
		StepStats: map[string]StepStat{
			"implement": {AvgCostUSD: 1.00, AvgDurationMS: 5000, FailRate: 0.10},
		},
		LintCatchRate:  0.25,
		ReviewFailRate: 0.15,
		ErrorBreakdown: map[string]int{"build_error": 3},
	}

	result := FormatMetrics(m)

	assert.Contains(t, result, "Total tasks: 30")
	assert.Contains(t, result, "Completed: 25")
	assert.Contains(t, result, "Failed: 5")
	assert.Contains(t, result, "83.3%")
	assert.Contains(t, result, "$45.00")
	assert.Contains(t, result, "sonnet")
	assert.Contains(t, result, "implement")
	assert.Contains(t, result, "25.0%")
	assert.Contains(t, result, "15.0%")
	assert.Contains(t, result, "build_error")
}

func TestProcessAdvisor_ParseOutput(t *testing.T) {
	raw := json.RawMessage(`{
		"insights": [
			{
				"title": "Reduce retry rate for lint-heavy tasks",
				"body": "Tasks with lint phase fail 40% of the time",
				"recommendation": "Run lint before implement to catch issues early",
				"category": "retry_rate",
				"priority": "high"
			}
		]
	}`)

	defs, err := parseProcessAdvisorOutput(raw)
	require.NoError(t, err)
	assert.Len(t, defs, 1)
	assert.Equal(t, "Reduce retry rate for lint-heavy tasks", defs[0].Title)
	assert.Equal(t, "retry_rate", defs[0].Category)
	assert.Equal(t, "high", defs[0].Priority)
}

func TestProcessAdvisor_ParseOutputEmpty(t *testing.T) {
	_, err := parseProcessAdvisorOutput(nil)
	assert.Error(t, err)
}

func TestProcessAdvisor_MapInsightCategory(t *testing.T) {
	assert.Equal(t, entities.InsightRetryRate, mapInsightCategory("retry_rate"))
	assert.Equal(t, entities.InsightCostOptimization, mapInsightCategory("cost_optimization"))
	assert.Equal(t, entities.InsightModelSelection, mapInsightCategory("model_selection"))
	assert.Equal(t, entities.InsightSpecQuality, mapInsightCategory("spec_quality"))
	assert.Equal(t, entities.InsightGateEffectiveness, mapInsightCategory("gate_effectiveness"))
	assert.Equal(t, entities.InsightWorkflowOptimization, mapInsightCategory("workflow_optimization"))
	assert.Equal(t, entities.InsightWorkflowOptimization, mapInsightCategory("unknown_category"))
}

func TestProcessAdvisor_CategorizeError(t *testing.T) {
	assert.Equal(t, "build_error", categorizeError("build failed"))
	assert.Equal(t, "test_error", categorizeError("test suite failed"))
	assert.Equal(t, "lint_error", categorizeError("lint check failed"))
	assert.Equal(t, "timeout", categorizeError("timeout exceeded"))
	assert.Equal(t, "budget_exceeded", categorizeError("budget limit reached"))
	assert.Equal(t, "other", categorizeError("unknown problem"))
}
