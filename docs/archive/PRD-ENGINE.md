# PRD: YOLO Factory — Autonomous Software Development Engine

## Problem Statement

The current YOLO Factory is a basic task runner with a monolithic execution job, hardcoded git operations, scattered retry logic, and no review system. It can run one Claude CLI command per task but has no workflow orchestration, no quality gates, no self-healing, and no structured context management.

Building YOLO (framework, plugins, and apps) requires executing dozens of PRDs with hundreds of tasks. Doing this interactively with Claude Code costs 5-8 hours per PRD. An autonomous system that takes a PRD and delivers finished, tested, reviewed code would save ~80% of that time.

## Solution

Rebuild Factory as a full autonomous development engine:

1. **PRD-driven**: Submit a PRD with acceptance criteria. Factory breaks it into tasks, executes them sequentially, reviews each one, and merges on success.
2. **Multi-step workflow per task**: Plan (Opus) → Implement/TDD (Sonnet) → Test (shell) → Audit (Sonnet) → Review (Sonnet). Each step is a fresh Claude Code session.
3. **Cross-project dependencies**: Tasks can depend on tasks in other projects. Factory executes in topological order.
4. **Self-healing**: Sentinel agent watches for broken builds, failing tests, security vulns. Auto-creates fix tasks.
5. **Suggestions**: Advisor agent analyzes code and execution history, suggests optimizations and refactoring.
6. **Full observability**: Track cost, tokens, duration per step/run/task/PRD. YAML backup to git repo.

## User Stories

1. As a developer, I want to submit a PRD and have Factory plan, implement, test, and merge all the work, so that I can focus on design and review instead of mechanical coding.
2. As a developer, I want Factory to break my PRD into ordered tasks with dependencies, so that complex features are implemented in the correct sequence.
3. As a developer, I want each task reviewed by a separate agent (not the one that wrote the code), so that quality issues are caught before merging.
4. As a developer, I want Factory to auto-retry failed tasks with error feedback and model escalation, so that transient failures are handled without my intervention.
5. As a developer, I want to manage multiple projects (core, plugins, apps) in Factory, so that I can queue work across the whole YOLO ecosystem.
6. As a developer, I want cross-project task dependencies, so that a plugin task waits for the core task it depends on.
7. As a developer, I want Factory to track cost and tokens per step, run, task, and PRD, so that I can understand and optimize spending.
8. As a developer, I want a Sentinel agent that auto-detects broken builds, failing tests, and security vulnerabilities, so that problems are caught and fixed without me checking manually.
9. As a developer, I want an Advisor agent that suggests optimizations, pattern extractions, and refactoring, so that code quality improves over time.
10. As a developer, I want all Factory state backed up as YAML in a git repo, so that I can recover from DB failures and have a full audit trail.
11. As a developer, I want to answer agent questions via CLI, admin UI, or MCP, so that blocked tasks can resume quickly.
12. As a developer, I want Factory to expose MCP tools, so that I can interact with Factory from interactive Claude Code sessions.
13. As a developer, I want configurable notifications (slack, email, webhook) for task failures, completions, and budget warnings.
14. As a developer, I want to pause, resume, and cancel projects and PRDs, so that I can control execution when priorities change.
15. As a developer, I want Factory to work with versioned branches (main, v1.x, v2.x), so that maintenance work and new development coexist.
16. As a developer, I want a CLI with full control over Factory (projects, PRDs, tasks, questions, suggestions, status, costs), so that I don't need the admin UI for basic operations.
17. As a developer, I want the admin UI to show a dashboard with running tasks, queue, costs, open questions, and suggestions.
18. As a developer, I want auto-merge by default (configurable), so that successful tasks flow to the target branch without manual PR review.
19. As a developer, I want model selection per workflow step (Opus for planning, Sonnet for implementation), so that cost is optimized without sacrificing quality.
20. As a developer, I want Factory to store task summaries and pass them as context to downstream tasks, so that agents understand what previous tasks did.

## YOLO Design

### Entities

**Project** — `server/factory/entities/project.go`
```
embed entity.BaseEntity
name              string    (unique, required)
status            string    (active|paused|archived, default: active)
repo_url          string    (required)
local_path        string    (required)
use_worktrees     bool      (default: false)
default_branch    string    (default: "main")
maintenance_branches json   (default: "[]")
default_model     string    (sonnet|opus|haiku, default: sonnet)
escalation_model  string    (sonnet|opus, default: opus)
escalation_after_retries int (default: 2)
budget_per_task_usd   float64 (default: 2.0)
budget_per_prd_usd    float64 (default: 20.0)
budget_monthly_usd    float64 (default: 200.0)
budget_warning_at     float64 (default: 0.8)
spent_this_month_usd  float64 (default: 0.0)
max_retries       int       (default: 3)
timeout_secs      int       (default: 600)
auto_merge        bool      (default: true)
auto_start        bool      (default: false)
push_failed_branches bool   (default: false)
setup_commands    json      (default: "[]")
test_commands     json      (default: '["go build ./...", "go test ./..."]')

relations:
  prds        has_many PRD
  tasks       has_many Task
  suggestions has_many Suggestion
```

**PRD** — `server/factory/entities/prd.go`
```
embed entity.BaseEntity
project_id         string   (FK, required)
title              string   (required)
status             string   (draft|approved|planning|in_progress|completed|failed, default: draft)
source             string   (manual|grill_me|factory_generated|imported, default: manual)
created_by         string   (human|advisor|sentinel, default: human)
body               text     (required)
acceptance_criteria json    (required, [{id, description, verification}])
design_decisions   json     (default: "[]")
total_tasks        int      (default: 0)
completed_tasks    int      (default: 0)
failed_tasks       int      (default: 0)
total_cost_usd     float64  (default: 0.0)
approved_at        timestamp (nullable)
completed_at       timestamp (nullable)

relations:
  project  belongs_to Project
  tasks    has_many Task
```

**Task** — `server/factory/entities/task.go`
```
embed entity.BaseEntity
prd_id              string  (FK, required)
project_id          string  (FK, required)
title               string  (required)
status              string  (queued|blocked|running|reviewing|done|failed|cancelled, default: queued)
spec                text    (required)
acceptance_criteria json    (required, [{id, description}])
branch              string  (required)
model               string  (nullable, override project default)
sequence            int     (required)
depends_on          json    (default: "[]", cross-project: "project-name:task-id")
run_count           int     (default: 0)
max_retries         int     (default: 3)
cost_usd            float64 (default: 0.0)
summary             text    (nullable, filled on completion)
commit_hash         string  (nullable)
started_at          timestamp (nullable)
completed_at        timestamp (nullable)

relations:
  prd       belongs_to PRD
  project   belongs_to Project
  runs      has_many Run
  questions has_many Question
```

**Run** — `server/factory/entities/run.go`
```
embed entity.BaseEntity
task_id          string    (FK, required)
agent_type       string    (planner|implementer|reviewer|auditor|sentinel|advisor, required)
status           string    (running|completed|failed|cancelled, default: running)
model            string    (required)
session_id       string    (nullable, Claude Code session UUID)
session_name     string    (nullable)
escalated_model  string    (nullable)
cost_usd         float64   (default: 0.0)
tokens_in        int       (default: 0)
tokens_out       int       (default: 0)
duration_ms      int       (default: 0)
num_turns        int       (default: 0)
commit_hash      string    (nullable)
branch_name      string    (nullable)
files_changed    json      (default: "[]")
result           text      (nullable)
error            text      (nullable)
started_at       timestamp (auto)
completed_at     timestamp (nullable)

relations:
  task    belongs_to Task
  steps   has_many Step
  review  has_one Review
```

**Step** — `server/factory/entities/step.go`
```
embed entity.BaseEntity
run_id       string  (FK, required)
phase        string  (plan|implement|test|audit|review, required)
skill        string  (required)
status       string  (running|completed|failed|skipped, default: running)
model        string  (required)
session_id   string  (nullable)
cost_usd     float64 (default: 0.0)
tokens_in    int     (default: 0)
tokens_out   int     (default: 0)
duration_ms  int     (default: 0)
input_summary  text  (nullable)
output_summary text  (nullable)
started_at   timestamp (auto)
completed_at timestamp (nullable)

relations:
  run  belongs_to Run
```

**Review** — `server/factory/entities/review.go`
```
embed entity.BaseEntity
run_id            string  (FK, required)
task_id           string  (FK, required)
session_id        string  (nullable)
model             string  (required)
verdict           string  (pass|fail, required)
reasons           json    (default: "[]")
anti_patterns     json    (default: "[]")
criteria_results  json    (required, [{criteria_id, passed, comment}])
suggestions       json    (default: "[]")
cost_usd          float64 (default: 0.0)

relations:
  run  belongs_to Run
  task belongs_to Task
```

**Question** — `server/factory/entities/question.go`
```
embed entity.BaseEntity
task_id            string  (FK, required)
run_id             string  (FK, required)
body               text    (required)
context            text    (nullable)
confidence         string  (low|medium, required)
status             string  (open|answered|auto_resolved, default: open)
answer             text    (nullable)
answered_by        string  (human|planner|auto, nullable)
answer_session_id  string  (nullable)
answered_at        timestamp (nullable)

relations:
  task belongs_to Task
  run  belongs_to Run
```

**Suggestion** — `server/factory/entities/suggestion.go`
```
embed entity.BaseEntity
project_id         string  (FK, required)
source             string  (sentinel|advisor, required)
category           string  (optimization|refactoring|tech_debt|security|new_feature|pattern_extraction|bug_fix, required)
title              string  (required)
body               text    (required)
priority           string  (low|medium|high|critical, default: medium)
status             string  (pending|approved|rejected|converted, default: pending)
converted_task_id  string  (FK, nullable)

relations:
  project        belongs_to Project
  converted_task belongs_to Task (nullable)
```

### Actions

**Project actions:**
- `CreateProjectAction` — `BaseCreate[Project, CreateProjectInput]`
- `UpdateProjectAction` — `BaseUpdate[Project, UpdateProjectInput]`
- `PauseProjectAction` — `NoInput` — sets status to paused, stops execution
- `ResumeProjectAction` — `NoInput` — sets status to active, resumes execution

**PRD actions:**
- `SubmitPRDAction` — `TypedInput[SubmitPRDInput]` — creates PRD in draft status
- `ApprovePRDAction` — `NoInput` — status draft→approved, triggers planning if auto_start
- `ExecutePRDAction` — `NoInput` — triggers task planning (Planner agent) then execution

**Task actions:**
- `CancelTaskAction` — `NoInput` — sets status to cancelled
- `RetryTaskAction` — `TypedInput[RetryTaskInput]` — requeue failed task, optional model override

**Run actions:**
- `CompleteRunAction` — `TypedInput[CompleteRunInput]` — maps run status→task status, unblocks dependents, handles retries

**Question actions:**
- `AnswerQuestionAction` — `TypedInput[AnswerQuestionInput]` — stores answer, resumes task execution

**Suggestion actions:**
- `ApproveSuggestionAction` — `TypedInput[ApproveSuggestionInput]` — converts to task in specified PRD
- `RejectSuggestionAction` — `TypedInput[RejectSuggestionInput]` — marks rejected with reason

### Response Design

All actions use `actx.Resolve("Entity", id)` + `action.OK()`. No direct entity data returns.

Extras needed:
- `ExecutePRDAction` → `action.OK(Extras{"task_count": n})` — how many tasks were generated
- `CompleteRunAction` → `action.OK(Extras{"next_task_id": id})` — what runs next

### Inputs

One file per action in `inputs/` directory:
- `CreateProjectInput` — name, repo_url, local_path, default_branch, default_model, budget fields
- `UpdateProjectInput` — all pointer fields for partial update
- `SubmitPRDInput` — project_id, title, body, acceptance_criteria, design_decisions, source
- `RetryTaskInput` — model (optional override)
- `CompleteRunInput` — status, cost_usd, tokens_in, tokens_out, duration_ms, num_turns, error, commit_hash, files_changed, result, session_id
- `AnswerQuestionInput` — answer
- `ApproveSuggestionInput` — prd_id (add to existing PRD or create new)
- `RejectSuggestionInput` — reason

### Filters

Registered via `registry.RegisterFilter()`:
- `ProjectFilter` — status, name
- `PRDFilter` — project_id, status, source, created_by
- `TaskFilter` — prd_id, project_id, status, branch, sequence
- `RunFilter` — task_id, status, agent_type, model
- `StepFilter` — run_id, phase, status
- `QuestionFilter` — task_id, status, confidence
- `SuggestionFilter` — project_id, source, category, status, priority

### Services (YOLO service pattern — `service.Base` embed, single `Execute` method)

**OrchestratorService** — `server/factory/services/orchestrator.go`
```go
type OrchestratorService struct {
    service.Base
    Claude    *claude.Client    // injected
    Git       *GitService       // injected
    Context   *ContextService   // injected
    // + read/write repos for Task, Run, Step, Review
}

type OrchestratorInput struct {
    TaskID string
}

type OrchestratorOutput struct {
    RunID     string
    Status    string  // completed | failed
    CostUSD   float64
}

func (s *OrchestratorService) Execute(ctx context.Context, input OrchestratorInput) (OrchestratorOutput, error)
// Internally: plan → implement → test → audit → review → merge
// Each step spawns claude CLI via s.Claude
// Handles retries, question escalation, model escalation
```

**PlannerService** — `server/factory/services/planner.go`
```go
type PlannerService struct {
    service.Base
    Claude *claude.Client
    // + read repos for PRD, Project
    // + write repo for Task
}

type PlannerInput struct {
    PRDID string
}

type PlannerOutput struct {
    TaskIDs []string
    Count   int
}

func (s *PlannerService) Execute(ctx context.Context, input PlannerInput) (PlannerOutput, error)
// Spawns Planner agent (Opus) to break PRD into tasks
// Creates Task entities with dependencies
// Validates: no cycles, all deps exist
```

**DependencyService** — `server/factory/services/dependency.go`
```go
type DependencyService struct {
    service.Base
    // + read repo for Task
}

type DependencyInput struct {
    TaskID    string
    DependsOn []string
}

type DependencyOutput struct {
    Valid        bool
    CycleError   string
    BlockedBy    []string
}

func (s *DependencyService) Execute(ctx context.Context, input DependencyInput) (DependencyOutput, error)
// Cycle detection (DFS), dependency validation, cross-project resolution
```

**ContextService** — `server/factory/services/context.go`
```go
type ContextService struct {
    service.Base
    // + read repos for Task, PRD, Project
}

type ContextInput struct {
    TaskID string
    Phase  string  // plan, implement, review, etc.
}

type ContextOutput struct {
    Prompt       string
    SystemPrompt string
}

func (s *ContextService) Execute(ctx context.Context, input ContextInput) (ContextOutput, error)
// Builds prompts from templates + task data + dependency summaries
// Different templates per phase
```

**GitService** — `server/factory/services/git.go`
```go
type GitService struct {
    service.Base
}

type GitInput struct {
    Operation  string  // checkout, branch, merge, push, pull, worktree_add, worktree_remove
    RepoPath   string
    Branch     string
    TaskID     string
}

type GitOutput struct {
    CommitHash    string
    FilesChanged  []string
    BranchName    string
}

func (s *GitService) Execute(ctx context.Context, input GitInput) (GitOutput, error)
// All git operations: branch, merge, push, pull, worktree management
```

**BackupService** — `server/factory/services/backup.go`
```go
type BackupService struct {
    service.Base
    StatePath string  // factory-state repo path
}

type BackupInput struct {
    Trigger  string   // task_change, prd_change, daily_snapshot, manual
    EntityType string // project, prd, task, question, suggestion
    EntityID   string
}

type BackupOutput struct {
    FilePath string
    CommitHash string
}

func (s *BackupService) Execute(ctx context.Context, input BackupInput) (BackupOutput, error)
// Marshal entity to YAML, write to file, git commit + push
```

**SentinelService** — `server/factory/services/sentinel.go`
```go
type SentinelService struct {
    service.Base
    Claude *claude.Client
    // + read/write repos for Project, Suggestion, Task
}

type SentinelInput struct {
    ProjectID string
    Watches   []string  // build_health, test_health, security, convention_drift, etc.
}

type SentinelOutput struct {
    Findings    []Finding
    TasksCreated int
    SuggestionsCreated int
}

func (s *SentinelService) Execute(ctx context.Context, input SentinelInput) (SentinelOutput, error)
// Runs health checks, creates fix tasks (high trust) or suggestions (medium/low trust)
```

**AdvisorService** — `server/factory/services/advisor.go`
```go
type AdvisorService struct {
    service.Base
    Claude *claude.Client
    // + read repos for Project, Run, Task
    // + write repo for Suggestion
}

type AdvisorInput struct {
    ProjectID    string
    AnalysisType string  // pattern_extraction, code_quality, performance, architecture, model_optimization
}

type AdvisorOutput struct {
    Suggestions []Suggestion
}

func (s *AdvisorService) Execute(ctx context.Context, input AdvisorInput) (AdvisorOutput, error)
// Analyzes code and execution history, creates suggestions
```

**QuestionResolverService** — `server/factory/services/question_resolver.go`
```go
type QuestionResolverService struct {
    service.Base
    Claude *claude.Client
    // + read repos for Question, Task
}

type QuestionResolverInput struct {
    QuestionID string
}

type QuestionResolverOutput struct {
    Answer     string
    AnsweredBy string  // auto, planner, (or remains open for human)
    Resolved   bool
}

func (s *QuestionResolverService) Execute(ctx context.Context, input QuestionResolverInput) (QuestionResolverOutput, error)
// Auto-resolve → Planner agent → leave for human
```

### Core Package: `claude` — `yolo/core/pkg/claude/`

```go
// client.go
type Client struct {
    CLIPath string  // path to claude binary, default: "claude"
}

type Config struct {
    Model          string
    AllowedTools   []string
    BudgetUSD      float64
    PermissionMode string
    Bare           bool
    Effort         string
    CWD            string
    SystemPrompt   string
    SessionName    string
    ResumeSession  string
    ForkSession    bool
    JSONSchema     string
    Env            []string
    Timeout        time.Duration
}

type Result struct {
    Result           string
    SessionID        string
    CostUSD          float64
    InputTokens      int
    OutputTokens     int
    DurationMS       int
    NumTurns         int
    IsError          bool
    StopReason       string
    StructuredOutput json.RawMessage
}

func (c *Client) Run(ctx context.Context, cfg Config, prompt string) (*Result, error)
func (c *Client) Resume(ctx context.Context, sessionID string, cfg Config, prompt string) (*Result, error)
func (c *Client) Fork(ctx context.Context, sessionID string, cfg Config, prompt string) (*Result, error)
```

### Jobs

All jobs call their corresponding service's `Execute` method:

- `ExecuteWorkflowJob` → calls `OrchestratorService.Execute()`
- `PlanPRDJob` → calls `PlannerService.Execute()`
- `CheckTimeoutsJob` → finds orphaned runs, marks failed, triggers retry
- `SentinelJob` → calls `SentinelService.Execute()` per project
- `AdvisorJob` → calls `AdvisorService.Execute()` per project
- `BackupSnapshotJob` → calls `BackupService.Execute()` with daily_snapshot trigger

### Commands (CLI)

```
factory project add|list|get|update|pause|resume|archive
factory prd submit|list|get|approve|execute|generate
factory task list|get|cancel|retry
factory status [--watch]
factory cost [--period week|month] [--project X]
factory questions list|answer
factory suggestions list|approve|reject
factory sentinel run [--project X] [--all]
factory advisor run [--project X] [--analysis X]
factory backup [--from PATH]
```

### Domain Structure

```
server/factory/
├── entities/    (8 entity files)
├── actions/     (12 action files)
├── inputs/      (8 input files)
├── filters/     (7 filter files)
├── services/    (8 service files)
├── jobs/        (6 job files)
├── skills/      (5 headless skill templates)
└── commands/    (10 command files)
```

### Plugin Integration

- **plugin-notifications**: Factory emits events via YOLO event system. Notification plugin delivers to configured channels.
- **core/pkg/claude**: Claude Code CLI wrapper as core package. Factory imports it. Other YOLO apps can too.

### app.yml Changes

Full app.yml documented in `docs/SRS.md` section 15.1. Key additions vs current:
- `agent_profiles` section (6 agent type configs)
- `backup` section (state repo path, auto-backup toggle)
- `sentinel` section (enabled, quick_checks_after_task)
- `advisor` section (enabled)
- `notifications` section (channels + event filters)
- Expanded `schedule` (sentinel-daily, sentinel-weekly, advisor-weekly, advisor-monthly, daily-backup)
- New `execution` queue for task workflow jobs

## Implementation Decisions

### Modules Built

1. **core/pkg/claude/** (NEW in yolo core) — Go wrapper for Claude Code CLI
2. **server/factory/entities/** (REWRITE) — 8 entities replacing current 4
3. **server/factory/actions/** (REWRITE) — 12 actions replacing current 10
4. **server/factory/services/** (NEW) — 8 services using YOLO service pattern
5. **server/factory/jobs/** (REWRITE) — 6 jobs replacing current 2
6. **server/factory/skills/** (NEW) — 5 headless prompt templates
7. **server/factory/commands/** (REWRITE) — 10 command groups replacing current 2
8. **server/factory/filters/** (REWRITE) — 7 filters replacing current 3
9. **migrations/** (REWRITE) — 9 migrations replacing current 4
10. **config/entities/** (REWRITE) — 8 UI YAML files replacing current 4

### Interfaces Modified

- `service.Service[I, O]` — used by all 8 Factory services (no modification needed, just usage)
- `entity.BaseEntity` — embedded by all 8 entities (no modification needed)

### Architectural Decisions

1. **Service-per-concern**: Each service has one responsibility and one `Execute` method
2. **Jobs are thin**: Jobs load input, call `service.Execute()`, handle result. No business logic in jobs.
3. **Skills are templates**: Go files that build prompts from templates + data. Not Claude Code plugin skills.
4. **Claude client injected**: All services receive `*claude.Client` via DI. Tests inject a mock client.
5. **Backup is async**: State changes write to DB (sync) then queue backup job (async). No blocking on git push.

### Schema Changes

Complete rewrite of all tables. Old tables (repos, tasks, runs, questions) are dropped. New tables:
- factory_projects
- factory_prds
- factory_tasks
- factory_runs
- factory_steps
- factory_reviews
- factory_questions
- factory_suggestions

Full migration SQL documented in `docs/SRS.md` section 16.

### State Machines

Documented in `docs/DESIGN.md` section 4:
- PRD: draft → approved → planning → in_progress → completed/failed
- Task: queued/blocked → running → reviewing → done/failed/cancelled
- Run: running → completed/failed/cancelled
- Question: open → answered/auto_resolved
- Suggestion: pending → approved/rejected → converted

## Testing Decisions

- **Integration tests only** — real DB, no mocks for repositories
- **Mock Claude CLI** — test helper creates a temporary script that returns configured JSON responses
- **Services are the test boundary** — test `OrchestratorService.Execute()` end-to-end with mock Claude CLI and real DB
- **E2E test** — full workflow: submit PRD → plan → execute → review → complete

### Modules to Test

| Module | Test Type | Priority |
|--------|-----------|----------|
| `core/pkg/claude` | Unit (mock subprocess) | High |
| `OrchestratorService` | Integration (mock CLI + real DB) | High |
| `PlannerService` | Integration (mock CLI + real DB) | High |
| `DependencyService` | Integration (real DB) | High |
| `ContextService` | Unit (template rendering) | Medium |
| `GitService` | Integration (real git repo) | Medium |
| `BackupService` | Integration (real filesystem) | Medium |
| `CompleteRunAction` | Integration (status transitions, unblocking) | High |
| `SentinelService` | Integration (mock CLI + real DB) | Medium |
| Full E2E workflow | E2E (mock CLI + real DB) | High |

## Out of Scope

- Parallel task execution within a project
- Cloud deployment (Docker, Kubernetes)
- Multi-language support (non-Go projects)
- PR-based workflow (Factory creates PRs instead of merging)
- GitHub issue import
- Custom agent types via plugin interface
- Cost prediction before execution
- A/B model testing
- Team support (multi-user, RBAC)
- TypeScript SDK sidecar
- Session forking for token optimization (future optimization)
- Auto-rollback on post-merge failures

## Further Notes

- Full SRS: `apps/factory/docs/SRS.md` (22 sections)
- Full Design: `apps/factory/docs/DESIGN.md` (9 sections with diagrams, templates, algorithms)
- PRD alignment review runs after ALL tasks in a PRD complete — catches scope drift
- Agent profiles are configurable in app.yml — model, tools, budget per agent type
- Notification events documented in SRS section 12 (12 event types)
- CLI commands documented in SRS section 13 (30+ commands)
- Admin UI views documented in SRS section 14 (11 views with dashboard layout)
