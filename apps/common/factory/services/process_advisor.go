package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// ProcessAdvisorService analyzes execution history and generates process improvement insights.
type ProcessAdvisorService struct {
	service.Base
	Claude       *claude.Client
	InsightWrite entity.WriteRepository[entities.Insight]
}

// ProcessAdvisorInput holds the data needed for process analysis.
type ProcessAdvisorInput struct {
	ProjectID         string
	MinCompletedTasks int // default: 20
}

// ProcessAdvisorOutput holds the generated insights.
type ProcessAdvisorOutput struct {
	Insights []entities.Insight
	Skipped  bool
	Reason   string
}

// ExecutionMetrics computed from input data.
type ExecutionMetrics struct {
	TotalTasks        int
	CompletedTasks    int
	FailedTasks       int
	SuccessRate       float64
	AvgRetriesPerTask float64
	TotalCostUSD      float64
	AvgCostPerTask    float64
	ModelStats        map[string]ModelStat
	StepStats         map[string]StepStat
	LintCatchRate     float64
	ReviewFailRate    float64
	ErrorBreakdown    map[string]int
}

// ModelStat holds per-model statistics.
type ModelStat struct {
	Tasks       int
	Successes   int
	AvgCostUSD  float64
	SuccessRate float64
}

// StepStat holds per-step-phase statistics.
type StepStat struct {
	AvgCostUSD    float64
	AvgDurationMS int
	FailRate      float64
}

// processAdvisorInsightDef is the structured output from the agent.
type processAdvisorInsightDef struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	Recommendation string `json:"recommendation"`
	Category       string `json:"category"`
	Priority       string `json:"priority"`
}

type processAdvisorOutput struct {
	Insights []processAdvisorInsightDef `json:"insights"`
}

// Execute runs the process advisor analysis.
func (s *ProcessAdvisorService) Execute(ctx context.Context, in ProcessAdvisorInput) (ProcessAdvisorOutput, error) {
	minTasks := in.MinCompletedTasks
	if minTasks == 0 {
		minTasks = 20
	}

	// 1. Load all data for the project.
	tasks, err := read.FindMany[entities.Task](ctx,
		read.Eq(fields.Task.ProjectID.Name(), in.ProjectID),
		read.Limit(1000),
	)
	if err != nil {
		return ProcessAdvisorOutput{}, fmt.Errorf("load tasks: %w", err)
	}

	runs, err := read.FindMany[entities.Run](ctx,
		read.OrderBy(fields.Run.CreatedAt.Name(), read.Desc),
		read.Limit(500),
	)
	if err != nil {
		return ProcessAdvisorOutput{}, fmt.Errorf("load runs: %w", err)
	}

	steps, err := read.FindMany[entities.Step](ctx, read.Limit(1000))
	if err != nil {
		return ProcessAdvisorOutput{}, fmt.Errorf("load steps: %w", err)
	}

	reviews, err := read.FindMany[entities.Review](ctx, read.Limit(500))
	if err != nil {
		return ProcessAdvisorOutput{}, fmt.Errorf("load reviews: %w", err)
	}

	lintResults, err := read.FindMany[entities.LintResult](ctx, read.Limit(500))
	if err != nil {
		return ProcessAdvisorOutput{}, fmt.Errorf("load lint results: %w", err)
	}

	// 2. Count completed tasks.
	completed := 0
	for _, t := range tasks {
		if t.Status == string(enums.TaskStatusDone) {
			completed++
		}
	}
	if completed < minTasks {
		return ProcessAdvisorOutput{
			Skipped: true,
			Reason:  fmt.Sprintf("only %d completed tasks, need %d", completed, minTasks),
		}, nil
	}

	// 3. Compute metrics.
	metrics := ComputeMetrics(tasks, runs, steps, reviews, lintResults)

	// 3. Format metrics summary.
	summary := FormatMetrics(metrics)

	// 4. Build prompt.
	prompt := strings.Replace(constants.ProcessAdvisorTemplate, "{{.MetricsSummary}}", summary, 1)

	// 5. Spawn Sonnet agent.
	result, err := s.Claude.Run(ctx, claude.Config{
		Model:          "sonnet",
		AllowedTools:   []string{},
		Bare:           true,
		BudgetUSD:      0.50,
		PermissionMode: "auto",
		Effort:         "medium",
		JSONSchema:     constants.ProcessAdvisorSchema,
		SessionName:    fmt.Sprintf("factory:project-%s:process-advisor", in.ProjectID),
		Timeout:        10 * time.Minute,
	}, prompt)
	if err != nil {
		return ProcessAdvisorOutput{}, fmt.Errorf("claude run: %w", err)
	}
	if result.IsError {
		return ProcessAdvisorOutput{}, fmt.Errorf("claude error: %s", result.Text)
	}

	// 6. Parse structured output.
	defs, err := parseProcessAdvisorOutput(result.StructuredOutput)
	if err != nil {
		return ProcessAdvisorOutput{}, fmt.Errorf("parse output: %w", err)
	}

	// 7. Convert to Insight entities.
	metricJSON, _ := json.Marshal(metrics)
	insights := make([]entities.Insight, len(defs))
	for i, def := range defs {
		insights[i] = entities.Insight{
			ProjectID:      in.ProjectID,
			Category:       mapInsightCategory(def.Category),
			Title:          def.Title,
			Body:           def.Body,
			Recommendation: def.Recommendation,
			Priority:       def.Priority,
			Status:         string(enums.InsightStatusPending),
			MetricData:     string(metricJSON),
		}
		insights[i].ID = ulid.Make().String()
	}

	// 8. Persist insights.
	for i := range insights {
		if _, err := s.InsightWrite.Insert(ctx, &insights[i]); err != nil {
			return ProcessAdvisorOutput{}, fmt.Errorf("insert insight: %w", err)
		}
	}

	return ProcessAdvisorOutput{Insights: insights}, nil
}

// ComputeMetrics computes execution metrics from raw data. Pure function, no DB queries.
func ComputeMetrics(tasks []entities.Task, runs []entities.Run, steps []entities.Step, reviews []entities.Review, lintResults []entities.LintResult) *ExecutionMetrics {
	m := &ExecutionMetrics{
		TotalTasks:     len(tasks),
		ModelStats:     make(map[string]ModelStat),
		StepStats:      make(map[string]StepStat),
		ErrorBreakdown: make(map[string]int),
	}

	// Task counts and retries.
	var totalRetries int
	for _, t := range tasks {
		switch t.Status {
		case string(enums.TaskStatusDone):
			m.CompletedTasks++
		case string(enums.TaskStatusFailed):
			m.FailedTasks++
		}
		totalRetries += t.RunCount
		m.TotalCostUSD += t.CostUSD
	}

	if m.TotalTasks > 0 {
		m.SuccessRate = float64(m.CompletedTasks) / float64(m.TotalTasks)
		m.AvgRetriesPerTask = float64(totalRetries) / float64(m.TotalTasks)
		m.AvgCostPerTask = m.TotalCostUSD / float64(m.TotalTasks)
	}

	// Model stats from runs.
	type modelAccum struct {
		tasks     int
		successes int
		totalCost float64
	}
	modelMap := make(map[string]*modelAccum)
	for _, r := range runs {
		model := r.Model
		if model == "" {
			model = "unknown"
		}
		acc, ok := modelMap[model]
		if !ok {
			acc = &modelAccum{}
			modelMap[model] = acc
		}
		acc.tasks++
		acc.totalCost += r.CostUSD
		if r.Status == string(enums.RunStatusCompleted) {
			acc.successes++
		}
	}
	for model, acc := range modelMap {
		stat := ModelStat{
			Tasks:     acc.tasks,
			Successes: acc.successes,
		}
		if acc.tasks > 0 {
			stat.AvgCostUSD = acc.totalCost / float64(acc.tasks)
			stat.SuccessRate = float64(acc.successes) / float64(acc.tasks)
		}
		m.ModelStats[model] = stat
	}

	// Step stats.
	type stepAccum struct {
		count      int
		totalCost  float64
		totalDurMS int
		failCount  int
	}
	stepMap := make(map[string]*stepAccum)
	for _, s := range steps {
		phase := s.Phase
		if phase == "" {
			phase = "unknown"
		}
		acc, ok := stepMap[phase]
		if !ok {
			acc = &stepAccum{}
			stepMap[phase] = acc
		}
		acc.count++
		acc.totalCost += s.CostUSD
		acc.totalDurMS += s.DurationMs
		if s.Status == string(enums.StepStatusFailed) {
			acc.failCount++
		}
	}
	for phase, acc := range stepMap {
		stat := StepStat{}
		if acc.count > 0 {
			stat.AvgCostUSD = acc.totalCost / float64(acc.count)
			stat.AvgDurationMS = acc.totalDurMS / acc.count
			stat.FailRate = float64(acc.failCount) / float64(acc.count)
		}
		m.StepStats[phase] = stat
	}

	// Lint catch rate: fraction of lint results that failed.
	if len(lintResults) > 0 {
		failed := 0
		for _, lr := range lintResults {
			if !lr.Passed {
				failed++
			}
		}
		m.LintCatchRate = float64(failed) / float64(len(lintResults))
	}

	// Review fail rate.
	if len(reviews) > 0 {
		failedReviews := 0
		for _, r := range reviews {
			if r.Verdict == string(enums.ReviewVerdictFail) {
				failedReviews++
			}
		}
		m.ReviewFailRate = float64(failedReviews) / float64(len(reviews))
	}

	// Error breakdown from failed runs.
	for _, r := range runs {
		if r.Status == string(enums.RunStatusFailed) && r.Error != "" {
			key := categorizeError(r.Error)
			m.ErrorBreakdown[key]++
		}
	}

	return m
}

// FormatMetrics formats metrics as a readable text summary.
func FormatMetrics(m *ExecutionMetrics) string {
	var b strings.Builder

	fmt.Fprintf(&b, "### Overview\n")
	fmt.Fprintf(&b, "- Total tasks: %d\n", m.TotalTasks)
	fmt.Fprintf(&b, "- Completed: %d\n", m.CompletedTasks)
	fmt.Fprintf(&b, "- Failed: %d\n", m.FailedTasks)
	fmt.Fprintf(&b, "- Success rate: %.1f%%\n", m.SuccessRate*100)
	fmt.Fprintf(&b, "- Avg retries/task: %.1f\n", m.AvgRetriesPerTask)
	fmt.Fprintf(&b, "- Total cost: $%.2f\n", m.TotalCostUSD)
	fmt.Fprintf(&b, "- Avg cost/task: $%.2f\n", m.AvgCostPerTask)

	if len(m.ModelStats) > 0 {
		fmt.Fprintf(&b, "\n### Model Stats\n")
		for model, stat := range m.ModelStats {
			fmt.Fprintf(&b, "- %s: %d runs, %d successes, %.1f%% success rate, $%.2f avg cost\n",
				model, stat.Tasks, stat.Successes, stat.SuccessRate*100, stat.AvgCostUSD)
		}
	}

	if len(m.StepStats) > 0 {
		fmt.Fprintf(&b, "\n### Step Stats\n")
		for phase, stat := range m.StepStats {
			fmt.Fprintf(&b, "- %s: $%.2f avg cost, %dms avg duration, %.1f%% fail rate\n",
				phase, stat.AvgCostUSD, stat.AvgDurationMS, stat.FailRate*100)
		}
	}

	fmt.Fprintf(&b, "\n### Quality Gates\n")
	fmt.Fprintf(&b, "- Lint catch rate: %.1f%%\n", m.LintCatchRate*100)
	fmt.Fprintf(&b, "- Review fail rate: %.1f%%\n", m.ReviewFailRate*100)

	if len(m.ErrorBreakdown) > 0 {
		fmt.Fprintf(&b, "\n### Error Breakdown\n")
		for errType, count := range m.ErrorBreakdown {
			fmt.Fprintf(&b, "- %s: %d\n", errType, count)
		}
	}

	return b.String()
}

// parseProcessAdvisorOutput parses the agent's structured JSON output.
func parseProcessAdvisorOutput(raw json.RawMessage) ([]processAdvisorInsightDef, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty structured output")
	}
	var out processAdvisorOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal insights: %w", err)
	}
	return out.Insights, nil
}

// mapInsightCategory validates and maps category strings.
func mapInsightCategory(category string) string {
	switch category {
	case string(enums.InsightCategoryRetryRate),
		string(enums.InsightCategoryCostOptimization),
		string(enums.InsightCategoryModelSelection),
		string(enums.InsightCategorySpecQuality),
		string(enums.InsightCategoryGateEffectiveness),
		string(enums.InsightCategoryWorkflowOptimization):
		return category
	default:
		return string(enums.InsightCategoryWorkflowOptimization)
	}
}

// categorizeError buckets error messages into broad categories.
func categorizeError(errMsg string) string {
	lower := strings.ToLower(errMsg)
	switch {
	case strings.Contains(lower, "build"):
		return "build_error"
	case strings.Contains(lower, "test"):
		return "test_error"
	case strings.Contains(lower, "lint"):
		return "lint_error"
	case strings.Contains(lower, "timeout"):
		return "timeout"
	case strings.Contains(lower, "budget"):
		return "budget_exceeded"
	default:
		return "other"
	}
}

func (s *ProcessAdvisorService) Description() string {
	return "Analyze factory process metrics and suggest improvements"
}
