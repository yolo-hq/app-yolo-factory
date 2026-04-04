# PRD: Factory Quality & Self-Improvement System

## Problem Statement

YOLO Factory v2 was built in 14 issues producing ~7,200 LOC. A post-implementation audit revealed:
- 4 critical bugs (command injection, swallowed errors, duplicate state machine, unfiltered queries)
- 15 code quality warnings (duplications, stubs, status mismatches, missing queue config)
- 14 planned features that were never implemented (budget enforcement, model escalation, notifications, etc.)

These issues exist because Factory's quality gates (audit, review, acceptance criteria verification) were designed in the SRS but never enforced during implementation. Each issue was implemented in isolation with no cross-cutting review, no acceptance criteria verification, and no program-level checks beyond `go build` and `go test`.

**The core problem:** Factory's workflow defines 5 steps (plan → implement → test → audit → review) but the test step only runs `go build` + `go test`, the audit step is a single agent pass, and there is no systematic check that planned features were actually implemented. 80% of the audit findings could have been caught by deterministic programs at zero token cost.

## Solution

A three-layer quality system that prevents the problems found in Factory's own development:

### Layer 1: Program Gates (deterministic, free, runs every task)
`factory lint` — a Go program that runs AST-based and grep-based checks. Catches code quality issues without spending tokens.

### Layer 2: Enhanced Agent Gates (costs tokens, runs every task)
Stricter review agent with per-criterion verification. Integration review every N tasks for cross-cutting concerns.

### Layer 3: Process Advisor (learns from execution history)
Analyzes retry rates, costs, model effectiveness, gate catch rates. Suggests process improvements. Needs 20+ completed tasks before producing insights.

**What this catches vs the audit:**
- Program gates catch: command injection, swallowed errors, duplicate functions, status string literals, stubs, missing queue config (C1, C2, W1-W3, W5-W6, W9-W11, W14)
- Agent gates catch: duplicate state machines, missing features, spec drift (C3, W7, all 14 missing features)
- Process Advisor catches: systemic patterns (specs too vague → high retry rate, wrong model selection, unnecessary gates)

## User Stories

1. As a developer, I want `factory lint` to catch code quality issues before the agent review step, so that I don't waste tokens on issues a program can detect.
2. As a developer, I want status strings enforced as constants (not string literals), so that status mismatches between entities are impossible.
3. As a developer, I want the review agent to verify each acceptance criterion individually with evidence, so that "done" means "all criteria actually met."
4. As a developer, I want an integration review every N tasks, so that cross-cutting concerns (shared helpers, consistent patterns) don't drift.
5. As a developer, I want PRD alignment review after all tasks complete, so that scope drift and missing features are caught before marking a PRD done.
6. As a developer, I want the Process Advisor to analyze execution history and suggest workflow improvements, so that Factory gets better over time.
7. As a developer, I want insights about model effectiveness (Opus vs Sonnet success rates per task type), so that I can optimize cost vs quality.
8. As a developer, I want insights about spec quality (tasks with vague specs retry more), so that I write better specs.
9. As a developer, I want insights about gate effectiveness (which gates catch real issues vs waste time), so that the pipeline stays lean.
10. As a developer, I want all quality check results stored and queryable, so that I can see trends over time.
11. As a developer, I want `factory lint` to run as part of the orchestrator's test step automatically, so that every task is checked.
12. As a developer, I want duplicate function detection across packages, so that shared helpers are extracted instead of copy-pasted.
13. As a developer, I want stub detection (functions with only fmt.Print body), so that incomplete implementations don't ship as "done."
14. As a developer, I want `_, _ =` pattern detection via AST, so that swallowed errors are caught at lint time.
15. As a developer, I want `sh -c` detection in exec.Command calls, so that command injection risks are caught at lint time.
16. As a developer, I want model escalation to work (Sonnet→Opus after N retries), so that hard tasks get more intelligence automatically.
17. As a developer, I want budget enforcement (per-task, per-PRD, monthly) with warnings, so that costs don't run away.
18. As a developer, I want question detection in agent output, so that confused agents pause instead of producing bad code.
19. As a developer, I want notifications emitted for task/PRD lifecycle events, so that I know what's happening without watching the CLI.
20. As a developer, I want scheduled jobs (timeout detection, monthly budget reset) working, so that the system self-maintains.
21. As a developer, I want auto_start working (PRD approved → auto-execute), so that I can submit PRDs and walk away.
22. As a developer, I want setup_commands run before first task in a project, so that new projects bootstrap automatically.

## YOLO Design

### Entities

**Insight** (new) — `server/factory/entities/insight.go`
```
embed entity.BaseEntity
project_id          string   (FK nullable — can be cross-project)
category            string   (retry_rate|cost_optimization|model_selection|spec_quality|gate_effectiveness|workflow_optimization)
title               string   (required)
body                text     (required)
metric_data         json     (required — raw numbers backing the insight)
recommendation      text     (required — actionable suggestion)
priority            string   (low|medium|high, default: medium)
status              string   (pending|acknowledged|applied|dismissed, default: pending)

relations:
  project  belongs_to Project (nullable)
```

**LintResult** (new) — `server/factory/entities/lint_result.go`
```
embed entity.BaseEntity
run_id              string   (FK, required)
task_id             string   (FK, required)
passed              bool     (required)
checks_run          int      (required)
checks_passed       int      (required)
checks_failed       int      (required)
findings            json     (required — [{check, severity, file, line, message}])

relations:
  run   belongs_to Run
  task  belongs_to Task
```

**Existing entity modifications:**
- Step: add `lint_result_id` field (nullable FK to LintResult)
- Run: add `lint_result_id` field (nullable FK to LintResult)

### Status Constants (new shared file)

`server/factory/entities/status.go`:
```go
package entities

// Project statuses
const (
    ProjectActive   = "active"
    ProjectPaused   = "paused"
    ProjectArchived = "archived"
)

// PRD statuses
const (
    PRDDraft      = "draft"
    PRDApproved   = "approved"
    PRDPlanning   = "planning"
    PRDInProgress = "in_progress"
    PRDCompleted  = "completed"
    PRDFailed     = "failed"
)

// Task statuses
const (
    TaskQueued    = "queued"
    TaskBlocked   = "blocked"
    TaskRunning   = "running"
    TaskReviewing = "reviewing"
    TaskDone      = "done"
    TaskFailed    = "failed"
    TaskCancelled = "cancelled"
)

// Run statuses
const (
    RunRunning   = "running"
    RunCompleted = "completed"
    RunFailed    = "failed"
    RunCancelled = "cancelled"
)

// Step statuses
const (
    StepRunning   = "running"
    StepCompleted = "completed"
    StepFailed    = "failed"
    StepSkipped   = "skipped"
)

// Review verdicts
const (
    ReviewPass = "pass"
    ReviewFail = "fail"
)

// Question statuses
const (
    QuestionOpen         = "open"
    QuestionAnswered     = "answered"
    QuestionAutoResolved = "auto_resolved"
)

// Suggestion statuses
const (
    SuggestionPending   = "pending"
    SuggestionApproved  = "approved"
    SuggestionRejected  = "rejected"
    SuggestionConverted = "converted"
)

// Insight statuses
const (
    InsightPending      = "pending"
    InsightAcknowledged = "acknowledged"
    InsightApplied      = "applied"
    InsightDismissed    = "dismissed"
)

// Agent types
const (
    AgentPlanner     = "planner"
    AgentImplementer = "implementer"
    AgentReviewer    = "reviewer"
    AgentAuditor     = "auditor"
    AgentSentinel    = "sentinel"
    AgentAdvisor     = "advisor"
)

// Step phases
const (
    PhasePlan      = "plan"
    PhaseImplement = "implement"
    PhaseTest      = "test"
    PhaseLint      = "lint"
    PhaseAudit     = "audit"
    PhaseReview    = "review"
)
```

### Actions

**Insight actions:**
- `AcknowledgeInsightAction` — NoInput, status pending→acknowledged
- `ApplyInsightAction` — NoInput, status acknowledged→applied
- `DismissInsightAction` — TypedInput[DismissInsightInput], status→dismissed with reason

### Inputs

- `DismissInsightInput` — reason (required)

### Filters

- `InsightFilter` — project_id, category, status, priority
- `LintResultFilter` — run_id, task_id, passed

### Domain structure

```
server/factory/
├── entities/
│   ├── status.go          (NEW — all status constants)
│   ├── insight.go         (NEW)
│   ├── lint_result.go     (NEW)
│   └── ... (existing, modified to use constants)
├── actions/
│   ├── insight_acknowledge.go  (NEW)
│   ├── insight_apply.go        (NEW)
│   ├── insight_dismiss.go      (NEW)
│   └── ... (existing, modified to use constants)
├── inputs/
│   └── insight_inputs.go  (NEW)
├── filters/
│   ├── insight_filter.go     (NEW)
│   ├── lint_result_filter.go (NEW)
│   └── ...
├── services/
│   ├── linter.go              (NEW — runs program checks)
│   ├── linter_test.go         (NEW)
│   ├── process_advisor.go     (NEW)
│   ├── process_advisor_test.go (NEW)
│   ├── helpers.go             (NEW — shared: truncate, parseDeps, toJSON)
│   └── ... (existing, modified to use constants + shared helpers)
├── jobs/
│   ├── check_timeouts.go     (NEW)
│   ├── reset_budgets.go      (NEW)
│   ├── process_advisor.go    (NEW)
│   └── ...
├── commands/
│   ├── lint.go                (NEW — factory lint command)
│   ├── insights.go            (NEW — factory insight list/acknowledge/apply/dismiss)
│   └── ...
└── lint/
    ├── checks.go              (NEW — all lint check implementations)
    ├── ast_checks.go          (NEW — Go AST-based checks)
    ├── grep_checks.go         (NEW — grep-based checks)
    └── checks_test.go         (NEW)
```

### Plugin integration

- **plugin-notifications**: Factory emits events at lifecycle points. Events already declared in `events/events.go`, need to be emitted in services/actions.
- **core/pkg/claude**: Already integrated. No changes.

### app.yml changes

```yaml
# Add to domains
factory:
  entities:
    Insight: "*"
    LintResult:
      queries: [list, get]

# Add to schedule
schedule:
  check-timeouts: "*/5 * * * *"
  reset-monthly-budgets: "0 0 1 * *"
  process-advisor: "0 5 * * 1"  # weekly Monday 5 AM

# Add to worker queues
queues:
  - critical
  - default
  - execution  # was missing
```

## Implementation Decisions

### Module 1: Shared Helpers (server/factory/services/helpers.go)
Extract duplicated functions: `truncate`, `parseDeps`, `toJSON`. Remove copies from orchestrator.go, run_complete.go, dependency.go.

**Tests:** Unit tests for each helper.

### Module 2: Status Constants (server/factory/entities/status.go)
Single source of truth. Replace ALL string literals across entire codebase. `grep -r '"queued"\|"done"\|"failed"'` should return zero matches outside status.go after this.

**Tests:** None — verified by compiler + grep check.

### Module 3: Factory Lint (server/factory/lint/)
Go program with two check types:

**AST checks** (parse Go source with `go/ast`):
- `check_swallowed_errors` — find `_, _ =` and `_ =` patterns on function calls that return error
- `check_no_shell_exec` — find `exec.Command("sh", "-c", ...)` patterns
- `check_entity_methods` — all structs embedding BaseEntity must have TableName() and EntityName()
- `check_resolve_before_ok` — actions calling action.OK() must have actx.Resolve() before it

**Grep checks** (search source text):
- `check_status_literals` — no hardcoded status strings outside status.go
- `check_duplicate_functions` — no exported functions with same name in different files
- `check_stub_functions` — functions where body is only fmt.Print/Println/Printf
- `check_todo_threshold` — fail if TODO/FIXME/HACK count exceeds configurable threshold

**Output:** LintResult entity with structured findings.

**Tests:** Test each check against sample Go code snippets.

### Module 4: Linter Service (server/factory/services/linter.go)
```go
type LinterService struct {
    service.Base
}

type LinterInput struct {
    ProjectPath string
    ChangedFiles []string  // scope to changed files for speed
}

type LinterOutput struct {
    Passed       bool
    ChecksRun    int
    ChecksPassed int
    ChecksFailed int
    Findings     []LintFinding
}

type LintFinding struct {
    Check    string  // check name
    Severity string  // error, warning, info
    File     string
    Line     int
    Message  string
}

func (s *LinterService) Execute(ctx context.Context, input LinterInput) (LinterOutput, error)
```

Integrated into orchestrator between test and audit steps:
```
plan → implement → test (go build+test+vet) → LINT (factory lint) → audit → review
```

**Tests:** Integration tests with sample Go files in testdata/.

### Module 5: Enhanced Review Agent
Modify review skill template to enforce per-criterion verification:
```
For each acceptance criterion:
1. State the criterion
2. Find the specific code/test that satisfies it
3. Quote the evidence (file:line)
4. Verdict: PASS or FAIL

If you cannot find evidence for a criterion, it FAILS.
Do not assume — show proof.
```

Update ReviewTaskSchema to require evidence field per criterion:
```json
{
  "criteria_results": [{
    "criteria_id": "tc-1",
    "passed": true,
    "evidence": "core/retry/policy.go:15 — RetryPolicy interface defined",
    "comment": "Interface exists with all required methods"
  }]
}
```

### Module 6: Integration Review
New service that runs every N tasks (configurable, default 5):

```go
type IntegrationReviewService struct {
    service.Base
    Claude *claude.Client
}

type IntegrationReviewInput struct {
    ProjectID    string
    TaskIDs      []string  // last N completed tasks
    AllDiffs     string    // combined git diffs
}
```

Checks:
- Duplicate functions across files (semantic, not just name)
- Inconsistent patterns (same problem solved differently in different files)
- Missing shared helpers (same logic repeated)
- State machine consistency (status transitions match across actions/services)

Creates Suggestion entities for issues found.

### Module 7: Process Advisor Service
```go
type ProcessAdvisorService struct {
    service.Base
    Claude *claude.Client
}

type ProcessAdvisorInput struct {
    MinCompletedTasks int  // default: 20, skip if not enough data
}

type ProcessAdvisorOutput struct {
    Insights []entities.Insight
    Skipped  bool    // true if insufficient data
    Reason   string  // why skipped
}
```

**Metrics computed from DB:**
```go
type ExecutionMetrics struct {
    // Per-project
    TaskCount          int
    SuccessRate        float64
    AvgRetriesPerTask  float64
    AvgCostPerTask     float64
    
    // Per-model
    ModelSuccessRates  map[string]float64  // sonnet: 0.85, opus: 0.95
    ModelAvgCosts      map[string]float64
    
    // Per-step
    StepFailRates      map[string]float64  // implement: 0.15, review: 0.08
    StepAvgCosts       map[string]float64
    
    // Per-gate
    LintCatchRate      float64  // % of tasks where lint found issues
    ReviewCatchRate    float64  // % of tasks where review failed
    AuditCatchRate     float64
    
    // Spec quality correlation
    AvgCriteriaPerTask float64
    SuccessRateByCriteriaCount map[int]float64  // 1 criteria: 60%, 3 criteria: 90%
    
    // Error categories
    ErrorBreakdown     map[string]int  // build_fail: 5, test_fail: 12, review_fail: 8
}
```

Agent prompt includes metrics summary, asks for insights with actionable recommendations.

**Tests:** Unit tests for metric computation. Mock Claude for insight generation.

### Module 8: Missing Feature Implementation
Not new architecture — just wiring what was designed but not connected:

- **Model escalation**: In orchestrator, check `run_count >= project.EscalationAfterRetries`, use `project.EscalationModel`
- **Budget enforcement**: In orchestrator before task start, check `project.SpentThisMonthUSD + project.BudgetPerTaskUSD <= project.BudgetMonthlyUSD`
- **Question detection**: In orchestrator after implement step, search result text for "QUESTION:" pattern, create Question entity, call QuestionResolverService
- **Event emission**: Add `events.Emit()` calls in orchestrator (TaskStarted, TaskCompleted, TaskFailed), CompleteRunAction (PRDCompleted, PRDFailed), sentinel (BuildBroken, SecurityVuln), question resolver (QuestionNeedsHuman)
- **Scheduled jobs**: CheckTimeoutsJob (find running runs past timeout, mark failed), ResetBudgetsJob (reset spent_this_month_usd monthly)
- **auto_start**: In ApprovePRDAction, if project.AutoStart, dispatch PlanPRDJob
- **push_failed_branches**: In orchestrator on failure, if project.PushFailedBranches, git push the branch
- **setup_commands**: In orchestrator before first task, check if setup has run, execute project.SetupCommands
- **Orphaned runs**: Add to sentinel watches
- **Wire CLI stubs**: Make sentinel:run, advisor:run, backup, recover dispatch actual jobs/services

### Module 9: Replace All String Literals
Mechanical find-and-replace across entire codebase:
```
"active"     → entities.ProjectActive
"queued"     → entities.TaskQueued
"done"       → entities.TaskDone
"completed"  → entities.RunCompleted
"failed"     → entities.TaskFailed / entities.RunFailed (context-dependent)
"running"    → entities.TaskRunning / entities.RunRunning
"draft"      → entities.PRDDraft
...etc
```

## Testing Decisions

- **Lint checks**: Test each check against crafted Go snippets in testdata/ — both passing and failing cases
- **Linter service**: Integration test — run lint on Factory's own codebase, verify expected findings
- **Process Advisor**: Unit test metric computation with synthetic data. Mock Claude for insight generation.
- **Integration Review**: Unit test for diff parsing. Mock Claude for review output.
- **Model escalation**: Integration test — mock Claude fail twice, verify third attempt uses Opus
- **Budget enforcement**: Integration test — set budget to $0.01, verify task fails with budget exceeded
- **Question detection**: Unit test — sample agent output with "QUESTION:", verify entity created
- **Status constants**: Grep test — verify zero string literal statuses outside status.go
- **All existing tests**: Must still pass after refactoring

## Out of Scope

- Custom lint rules via configuration (hardcoded checks for now)
- Web UI for Process Advisor insights (CLI + admin entity view is enough)
- Auto-applying Process Advisor recommendations (human reviews insights)
- Cross-repo lint checks (lint runs per-project)
- Performance profiling of Factory itself
- A/B testing of different models (manual model selection, advisor suggests)

## Further Notes

- `factory lint` should be fast (<5 seconds on Factory's own codebase). AST parsing is slower than grep but still sub-second per file.
- Process Advisor runs weekly by default. Can be triggered manually via `factory advisor:process`.
- Integration review frequency (every N tasks) is configurable in app.yml.
- The Insight entity has `metric_data` JSON field that stores the raw numbers backing each insight, so humans can verify the agent's reasoning.
- All string literal replacements must be done atomically (one commit) to avoid partial states.
- The lint step in the orchestrator is a PROGRAM step (like test), not an AGENT step. Zero token cost.
- Events are emitted via YOLO's event system. If plugin-notifications is not configured, events are logged but not delivered. No failure if notifications aren't set up.
