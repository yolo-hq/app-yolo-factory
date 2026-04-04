# YOLO Factory — Design Document

**Version:** 1.0.0
**Date:** 2026-04-04
**Status:** Draft

---

## 1. Design Decisions Log

Decisions made during the design process and the reasoning behind each.

### D1: Sequential Execution Per Project

**Decision:** One task at a time per project. Multiple projects can run in parallel.

**Why:** Parallel tasks within a project cause git conflicts, port conflicts, DB conflicts, and merge ordering problems. YOLO is a framework — almost every change touches shared core types. Sequential execution eliminates these problems entirely.

**Trade-off:** Slower per project. Mitigated by: better specs produce faster completions, and multiple projects run simultaneously since they're separate repos.

### D2: No GitHub Issues — Factory is Source of Truth

**Decision:** Factory tasks replace GitHub issues for automated work. PRs still go to GitHub.

**Why:** Two systems = sync problems. Factory tasks need structured data (deps, acceptance criteria, branch) that GitHub issues can't enforce. GitHub issues remain for external bug reports — Factory can import them.

**Trade-off:** Lose GitHub's free collaboration UI. Mitigated by: Factory admin UI provides the same visibility.

### D3: Go CLI Wrapper (Not TypeScript Sidecar)

**Decision:** Shell out to `claude` CLI from Go, wrapped in a typed Go package.

**Why:**
- No Go SDK exists for Claude Code. Only Python and TypeScript SDKs.
- CLI's `--output-format json` provides everything needed: cost, tokens, session_id, result.
- Sidecar adds a second runtime (Node/Bun), second process, communication overhead, and extra failure modes.
- `--resume` and `--fork-session` CLI flags handle multi-turn and context reuse.

**Trade-off:** Lose TypeScript SDK's `ClaudeSDKClient` multi-turn convenience. Mitigated by: `--resume` achieves the same result via CLI.

### D4: YOLO App (Not Standalone)

**Decision:** Factory is a YOLO app using YOLO framework patterns.

**Why:** Dogfooding — Factory proves YOLO works for real apps. Gets entities, actions, admin UI, MCP, worker jobs, CLI framework for free. The circular dependency concern (Factory builds YOLO but needs YOLO) is not real — Factory uses a released version to build the next version, like GCC compiling GCC.

### D5: Worktrees Optional, Default Off

**Decision:** Worktrees are configurable per project, defaulting to off.

**Why:** Local development often uses `go.work` with `replace` directives for linked modules. Worktrees break these relative paths. For standalone repos or cloud execution, worktrees provide safety (failed tasks don't touch main).

**Configuration:** `project.use_worktrees: bool`

### D6: YAML Backup in Dedicated Git Repo

**Decision:** All entity state is backed up as YAML files in a dedicated `factory-state` git repo, auto-committed on every state change.

**Why:**
- Human-readable (can `cat` a task file)
- Git-diffable (see exactly what changed and when)
- Portable (no binary format dependency)
- Full audit trail via git history
- Recovery: `factory recover --from` rebuilds DB from YAML files

**Trade-off:** Write overhead on every state change. Mitigated by: async git commit/push, and state changes are infrequent (seconds between them, not milliseconds).

### D7: Implementation Agent != Review Agent

**Decision:** The agent that writes code never reviews its own work. Reviews are always a fresh session with a different agent.

**Why:** From superpowers project research — "agent can't mark its own homework." A fresh context with read-only tools produces honest reviews. The implementation agent has context bias.

### D8: Structured Output via --json-schema

**Decision:** Factory skills return structured JSON, not prose markdown.

**Why:** Factory needs to parse agent output programmatically. `--json-schema` forces Claude Code to return valid JSON matching a schema. This eliminates fragile text parsing.

### D9: Model Per Step, Not Per Task

**Decision:** Each step within a run can use a different model. Plan uses Opus, Implement uses Sonnet, etc.

**Why:** Different phases need different intelligence levels. Planning requires creativity (Opus). Implementation follows patterns (Sonnet). Auditing is mechanical (Sonnet/Haiku). This optimizes cost without sacrificing quality where it matters.

### D10: Session Resume Between Plan and Implement Steps

**Decision:** Implementation step resumes the Plan step's session, not a fresh session.

**Why:** The plan is context the implementer needs. Resuming carries it over without re-injecting. Saves tokens and preserves nuance from the planning phase.

**Exception:** Review and Audit always use fresh sessions — they must be independent of the implementation context.

### D11: Question Escalation Chain

**Decision:** Agent questions go through: auto-resolve → Planner agent → human, in that order.

**Why:** Most questions can be answered by reading docs or asking a smarter model. Only truly novel questions need human input. This keeps Factory autonomous for routine work while having a human escape hatch.

### D12: Sentinel Auto-Creates, Advisor Suggests

**Decision:** Sentinel can auto-create tasks for high-trust issues (build broken, security vuln). Advisor only creates suggestions that humans approve.

**Why:** Build failures are objective — if `go build` fails, it's broken. Optimization suggestions are subjective — the human decides if they're worth pursuing. Different trust levels for different signal quality.

### D13: Task Summaries as Context

**Decision:** When a task completes, Factory stores a short summary. Downstream tasks receive these summaries as context.

**Why:** Downstream agents need to know what previous tasks did without reading the full conversation history. A 2-3 sentence summary is cheaper than re-analyzing git diffs.

### D14: No Sub-Tasks

**Decision:** Tasks don't have children. If a task is too big, split it into separate tasks at the PRD level.

**Why:** Sub-tasks add nested state machines (partial completion, nested dependencies). Steps within a run already handle the "phases within a task" concept. Adding sub-tasks creates two competing hierarchy systems.

**Rule:** If a task needs sub-tasks, the task is too big. Split it.

### D15: Auto-Merge by Default

**Decision:** When a task passes all gates (build, test, audit, review), Factory merges to the target branch and pushes.

**Why:** The whole point is autonomous execution. If every task waits for human merge approval, Factory becomes a fancy PR generator. Trust the automated gates.

**Override:** `project.auto_merge: false` creates PRs instead.

---

## 2. Architecture Diagrams

### 2.1 Data Flow

```
Human
  │
  │ PRD (markdown + acceptance criteria)
  ▼
┌─────────────┐
│ PRD Entity   │ status: draft → approved → planning → in_progress → completed
└──────┬──────┘
       │ Planner agent breaks into tasks
       ▼
┌─────────────┐
│ Task 1       │ status: queued → running → reviewing → done
│ Task 2       │ status: blocked (depends on Task 1) → queued → ...
│ Task 3       │ status: blocked (depends on Task 2) → queued → ...
└──────┬──────┘
       │ For each task (sequential):
       ▼
┌─────────────┐
│ Run          │ status: running → completed/failed
│  ├ Step: plan│ → implementation plan
│  ├ Step: impl│ → code changes
│  ├ Step: test│ → build + test verification
│  ├ Step: aud │ → convention check
│  └ Step: rev │ → acceptance verification
│              │
│ Review       │ verdict: pass/fail
└──────┬──────┘
       │ If pass:
       ▼
  Git merge + push
  Task status → done
  Unblock dependents
  Backup to factory-state
```

### 2.2 Agent Orchestration Flow

```
Factory Worker (Go)
  │
  ├── Compose prompt
  │     ├── Task spec
  │     ├── Acceptance criteria
  │     ├── Previous task summaries
  │     └── Framework context (CLAUDE.md excerpt)
  │
  ├── Select agent profile
  │     └── {model, tools, budget, bare, effort}
  │
  ├── Build CLI command
  │     └── claude -p --bare --output-format json --model X ...
  │
  ├── Spawn subprocess
  │     └── exec.Command("claude", args...)
  │
  ├── Wait for completion (with timeout)
  │     ├── Parse JSON result
  │     ├── Extract: cost, tokens, session_id, result, is_error
  │     └── Handle: question detection, error handling
  │
  ├── Update DB
  │     ├── Run: cost, tokens, duration, status
  │     ├── Step: cost, duration, output_summary
  │     └── Task: aggregate cost, status
  │
  └── Trigger next step or next task
```

### 2.3 Question Escalation Flow

```
Implementer agent outputs question
  │
  ▼
Factory detects question in result
  │
  ├── Step 1: Auto-resolve
  │   Search: CLAUDE.md, docs, previous task summaries
  │   Found? → Resume implementer with answer
  │
  ├── Step 2: Ask Planner (Opus)
  │   Spawn: fresh session with question + full context
  │   Got answer? → Resume implementer with --resume
  │
  └── Step 3: Ask Human
      Notify: via plugin-notifications
      Pause: task stays "running", waiting
      Human answers via CLI/UI/MCP
      Resume: implementer with --resume
```

### 2.4 Backup Flow

```
State change (task status, PRD update, etc.)
  │
  ├── Write to PostgreSQL (primary, sync)
  │
  └── Write to factory-state repo (backup, async)
        ├── Marshal entity to YAML
        ├── Write to file (tasks/task-001.yml)
        ├── git add + commit (message: "task-001: queued → running")
        └── git push (best-effort, retry on failure)

Daily (midnight):
  ├── Export full DB to snapshots/YYYY-MM-DD.yml
  ├── git add + commit + push
  └── Prune old snapshots (keep last 30 days)
```

---

## 3. Prompt Templates

### 3.1 Plan-Tasks Prompt

```
You are a software architect breaking a PRD into implementation tasks.

## Project
Name: {project.name}
Repository: {project.repo_url}
Branch: {task.branch}

## Framework Conventions
{claude_md_excerpt}

## PRD
Title: {prd.title}

{prd.body}

## Acceptance Criteria
{for each ac in prd.acceptance_criteria}
- [{ac.id}] {ac.description} (verify: {ac.verification})
{end}

## Design Decisions
{for each dd in prd.design_decisions}
- {dd}
{end}

## Instructions
Break this PRD into ordered tasks. Each task must:
1. Target ONE repository and ONE branch
2. Be independently testable (build + tests pass after just this task)
3. Have specific, verifiable acceptance criteria
4. List dependencies on other tasks by sequence number
5. Be small enough to complete in one agent session

Cross-project dependencies use format: "project-name:sequence"

Output the task list as structured JSON.
```

### 3.2 Implement Prompt

```
You are implementing a software task following TDD methodology.

## Task
Title: {task.title}

{task.spec}

## Acceptance Criteria
{for each ac in task.acceptance_criteria}
- [{ac.id}] {ac.description}
{end}

## Implementation Plan
{plan_step.output_summary}

## Previous Tasks Completed
{for each dep in completed_dependencies}
- {dep.title} (commit {dep.commit_hash}): {dep.summary}
  Changed: {dep.files_changed}
{end}

{if retry}
## Previous Attempt Failed
Error: {previous_run.error}
Review feedback: {previous_review.reasons}
Files changed in failed attempt: {previous_run.files_changed}

Address the specific errors above. Do not repeat the same approach.
{end}

## Instructions
1. Read existing code to understand current state
2. For each acceptance criterion:
   a. Write a failing test (red)
   b. Write minimal implementation to pass (green)
   c. Refactor if needed
3. Ensure ALL tests pass: go build ./... && go test ./...
4. Do not change code beyond what the task spec asks for
```

### 3.3 Review Prompt

```
You are reviewing a code implementation against its acceptance criteria.

## Task
Title: {task.title}
Spec: {task.spec}

## Acceptance Criteria
{for each ac in task.acceptance_criteria}
- [{ac.id}] {ac.description}
{end}

## Changes Made
{git_diff}

## Files Changed
{files_changed}

## Anti-Pattern Checklist
Check for:
- Hardcoded values that should be configurable
- Missing error handling at system boundaries
- Tests that mock internal code instead of using real implementations
- Code that violates YOLO entity/action patterns
- Scope creep — changes beyond what the task spec asked for
- Missing or incorrect type annotations
- Untested edge cases mentioned in the spec

## Instructions
Review the changes against each acceptance criterion. For each criterion, state whether it passes and why.

If you find anti-patterns or issues, list them.

Output your verdict as structured JSON: pass or fail with detailed reasons.
```

### 3.4 Review-PRD Prompt

```
You are performing a final alignment review of a completed PRD.

## Original PRD
Title: {prd.title}

{prd.body}

## PRD Acceptance Criteria
{for each ac in prd.acceptance_criteria}
- [{ac.id}] {ac.description} (verify: {ac.verification})
{end}

## Tasks Completed
{for each task in prd.tasks where status == "done"}
### {task.sequence}. {task.title}
Summary: {task.summary}
Commit: {task.commit_hash}
Files changed: {task.files_changed}
Review verdict: {task.review.verdict}
{end}

## Instructions
Compare what was requested in the PRD against what was actually delivered.

Check:
1. Are all acceptance criteria met?
2. Is there scope drift (things built that weren't asked for)?
3. Is there scope reduction (things asked for but not built)?
4. Do the tasks integrate correctly with each other?
5. Are there gaps between individual task completions and the overall PRD goal?

Output: alignment score (0.0-1.0), criteria met/missed, and recommendations.
```

### 3.5 Sentinel Prompt

```
You are a code health sentinel checking project {project.name}.

## Checks to Perform
{for each watch in watches}
- {watch.name}: {watch.description}
{end}

## Instructions
Run each check and report findings.

For critical issues (build broken, tests failing, security vulnerabilities):
  Create a task suggestion with category "bug_fix" and priority "critical"

For non-critical issues (convention drift, TODOs, outdated deps):
  Create a suggestion with appropriate category and priority

Output findings as structured JSON.
```

### 3.6 Advisor Prompt

```
You are an optimization advisor analyzing project {project.name}.

## Analysis Type: {analysis_type}

## Context
{analysis_specific_context}

## Run History (last 30 days)
Total runs: {run_count}
Success rate: {success_rate}%
Average cost per task: ${avg_cost}
Average retries per task: {avg_retries}
Most expensive tasks: {top_expensive}
Most retried tasks: {top_retried}

## Instructions
Analyze the project and suggest improvements.

Categories: optimization, refactoring, tech_debt, new_feature, pattern_extraction

For each suggestion:
- Title (short, actionable)
- Body (what to do and why)
- Priority (low/medium/high)
- Estimated impact

Output as structured JSON.
```

---

## 4. State Machines

### 4.1 PRD State Machine

```
           submit
  (none) ────────► draft
                     │
                     │ approve (human or auto)
                     ▼
                  approved
                     │
                     │ plan-tasks (Planner agent)
                     ▼
                  planning
                     │
                     │ tasks created, execution starts
                     ▼
                 in_progress
                   │     │
                   │     │ all tasks done
                   │     ▼
                   │  completed
                   │
                   │ any task failed after max retries
                   ▼
                  failed
```

### 4.2 Task State Machine

```
                 create
  (none) ─────────────────► queued (no deps) OR blocked (has unmet deps)

  blocked ──── all deps done ────► queued

  queued ───── picked for execution ────► running

  running ──── all steps pass ────► reviewing

  reviewing ── review passes ────► done
           └── review fails ─────► running (retry with feedback)

  running ──── step fails ────► running (retry) OR failed (max retries)

  done ────── (terminal state)
  failed ──── (terminal, but human can retry)
  cancelled ── (terminal, human cancelled)

  Any state ── human cancels ────► cancelled
```

### 4.3 Run State Machine

```
  create ────► running
                 │
                 ├── all steps complete, review passes ──► completed
                 ├── any step fails ─────────────────────► failed
                 ├── timeout exceeded ───────────────────► failed
                 ├── budget exceeded ────────────────────► failed
                 └── human cancels ──────────────────────► cancelled
```

### 4.4 Question State Machine

```
  create (agent raises) ────► open
                               │
                               ├── answer found in docs ──► auto_resolved
                               ├── planner answers ───────► answered
                               └── human answers ─────────► answered
```

### 4.5 Suggestion State Machine

```
  create (sentinel/advisor) ────► pending
                                    │
                                    ├── human approves ──► approved ──► converted (task created)
                                    └── human rejects ───► rejected
```

---

## 5. Token Optimization Strategy

### 5.1 Cost Estimates Per Task

```
Step           Model    Est. Input   Est. Output   Est. Cost
──────────────────────────────────────────────────────────────
Plan           Opus     ~20K tokens  ~2K tokens    ~$0.40
Implement      Sonnet   ~40K tokens  ~8K tokens    ~$0.30
Test           (shell)  N/A          N/A           $0.00
Audit          Sonnet   ~15K tokens  ~1K tokens    ~$0.08
Review         Sonnet   ~25K tokens  ~2K tokens    ~$0.15
──────────────────────────────────────────────────────────────
Total per task                                     ~$0.93
```

### 5.2 Optimization Techniques

| Technique | Estimated Savings | Implementation |
|-----------|------------------|----------------|
| `--bare` mode for all Factory agents | ~15K tokens/session | Agent wrapper sets `--bare` flag |
| `--tools` restriction per agent type | ~2K tokens/session | Only load tools the agent needs |
| `--effort low` for Audit | ~30% of audit cost | Agent profile config |
| Haiku for Sentinel | ~90% vs Sonnet | Agent profile config |
| Resume between Plan → Implement | ~30K tokens (no re-inject) | `--resume` with plan session_id |
| Compact task summaries (< 200 words) | ~5K tokens/task in context | Prompt template enforces brevity |
| Limit context to last 5 dep summaries | ~25K tokens max | Factory caps injected context |
| Skip audit if no convention-sensitive files changed | ~$0.08/task when applicable | Factory checks files_changed |

### 5.3 Monthly Cost Projection

```
Light usage (5 PRDs/month, 5 tasks each):
  Tasks: 25 × $0.93                    = $23.25
  PRD planning: 5 × $0.50              = $2.50
  PRD reviews: 5 × $0.20               = $1.00
  Retries (20%): 5 × $0.93             = $4.65
  Sentinel (daily): 30 × $0.20         = $6.00
  Advisor (weekly): 4 × $0.50          = $2.00
                               Total   ≈ $39.40/month

Heavy usage (15 PRDs/month, 7 tasks each):
  Tasks: 105 × $0.93                   = $97.65
  PRD planning: 15 × $0.50             = $7.50
  PRD reviews: 15 × $0.20              = $3.00
  Retries (20%): 21 × $0.93            = $19.53
  Sentinel (daily): 30 × $0.20         = $6.00
  Advisor (weekly): 4 × $0.50          = $2.00
                               Total   ≈ $135.68/month
```

---

## 6. Directory Structure

### 6.1 Factory App Structure

```
apps/factory/
├── CLAUDE.md                          # Framework conventions for this app
├── app.yml                            # Full configuration
├── main.go                            # Generated entry point
├── setup.go                           # Registration for tests
├── go.mod
├── go.sum
│
├── docs/
│   ├── SRS.md                         # Software Requirements Specification
│   └── DESIGN.md                      # This document
│
├── pkg/
│   └── claude/                        # Go wrapper for Claude Code CLI
│       ├── agent.go                   # AgentConfig, Run(), Resume(), Fork()
│       ├── result.go                  # AgentResult, JSON parsing
│       ├── stream.go                  # StreamResult, NDJSON parsing
│       └── session.go                 # Session management utilities
│
├── server/factory/
│   ├── entities/
│   │   ├── project.go
│   │   ├── prd.go
│   │   ├── task.go
│   │   ├── run.go
│   │   ├── step.go
│   │   ├── review.go
│   │   ├── question.go
│   │   └── suggestion.go
│   │
│   ├── actions/
│   │   ├── project_create.go
│   │   ├── project_update.go
│   │   ├── project_pause.go
│   │   ├── project_resume.go
│   │   ├── prd_submit.go
│   │   ├── prd_approve.go
│   │   ├── prd_execute.go            # triggers planning + execution
│   │   ├── task_execute.go            # picks next task, starts workflow
│   │   ├── task_cancel.go
│   │   ├── task_retry.go
│   │   ├── run_complete.go            # handles run completion, triggers next
│   │   ├── question_answer.go
│   │   ├── suggestion_approve.go
│   │   └── suggestion_reject.go
│   │
│   ├── inputs/
│   │   ├── project_inputs.go
│   │   ├── prd_inputs.go
│   │   ├── task_inputs.go
│   │   ├── run_inputs.go
│   │   ├── question_inputs.go
│   │   └── suggestion_inputs.go
│   │
│   ├── filters/
│   │   ├── project_filter.go
│   │   ├── prd_filter.go
│   │   ├── task_filter.go
│   │   ├── run_filter.go
│   │   ├── step_filter.go
│   │   ├── question_filter.go
│   │   └── suggestion_filter.go
│   │
│   ├── jobs/
│   │   ├── execute_workflow.go        # main task execution workflow
│   │   ├── plan_prd.go                # PRD → tasks planning job
│   │   ├── check_timeouts.go          # orphaned run detection
│   │   ├── sentinel.go                # health checks, security scans
│   │   ├── advisor.go                 # optimization analysis
│   │   ├── backup_state.go            # YAML backup to git repo
│   │   └── daily_snapshot.go          # full DB dump
│   │
│   ├── services/
│   │   ├── orchestrator.go            # task execution orchestration logic
│   │   ├── dependency.go              # dependency resolution + cycle detection
│   │   ├── context_builder.go         # builds prompts from templates + data
│   │   ├── git.go                     # git operations (branch, merge, push)
│   │   └── backup.go                  # YAML serialization + git backup
│   │
│   ├── skills/                        # headless skill prompt templates
│   │   ├── plan_tasks.go
│   │   ├── implement.go
│   │   ├── review_task.go
│   │   ├── review_prd.go
│   │   └── audit.go
│   │
│   └── commands/
│       ├── project.go                 # factory project add/list/get/update
│       ├── prd.go                     # factory prd submit/approve/execute
│       ├── task.go                    # factory task list/get/cancel/retry
│       ├── status.go                  # factory status [--watch]
│       ├── cost.go                    # factory cost --period --project
│       ├── questions.go               # factory questions list/answer
│       ├── suggestions.go             # factory suggestions list/approve/reject
│       ├── sentinel.go                # factory sentinel run
│       ├── advisor.go                 # factory advisor run
│       ├── backup.go                  # factory backup / factory recover
│       └── setup.go                   # factory setup (initial config)
│
├── migrations/
│   ├── 001_create_projects.up.sql
│   ├── 001_create_projects.down.sql
│   ├── 002_create_prds.up.sql
│   ├── 002_create_prds.down.sql
│   ├── 003_create_tasks.up.sql
│   ├── 003_create_tasks.down.sql
│   ├── 004_create_runs.up.sql
│   ├── 004_create_runs.down.sql
│   ├── 005_create_steps.up.sql
│   ├── 005_create_steps.down.sql
│   ├── 006_create_reviews.up.sql
│   ├── 006_create_reviews.down.sql
│   ├── 007_create_questions.up.sql
│   ├── 007_create_questions.down.sql
│   ├── 008_create_suggestions.up.sql
│   ├── 008_create_suggestions.down.sql
│   └── 009_create_indexes.up.sql
│
├── config/
│   ├── clients/admin.ui.yml
│   └── entities/factory/
│       ├── project.ui.yml
│       ├── prd.ui.yml
│       ├── task.ui.yml
│       ├── run.ui.yml
│       ├── step.ui.yml
│       ├── review.ui.yml
│       ├── question.ui.yml
│       └── suggestion.ui.yml
│
├── clients/admin/                     # React admin UI
│
└── e2e_test.go                        # End-to-end tests
```

### 6.2 Factory State Repository Structure

```
factory-state/
├── README.md                          # auto-generated, describes structure
├── projects/
│   ├── yolo-core.yml
│   ├── plugin-webhooks.yml
│   └── app-libromi.yml
├── prds/
│   ├── prd-001.yml
│   └── prd-002.yml
├── tasks/
│   ├── task-001.yml                   # includes nested runs/steps/reviews
│   ├── task-002.yml
│   └── task-003.yml
├── questions/
│   └── question-001.yml
├── suggestions/
│   └── suggestion-001.yml
└── snapshots/
    ├── 2026-04-04.yml
    └── 2026-04-05.yml
```

---

## 7. Key Algorithms

### 7.1 Dependency Cycle Detection

```go
// DFS-based cycle detection on task creation
func detectCycle(taskID string, dependsOn []string, allTasks map[string]*Task) error {
    visited := make(map[string]bool)
    path := make(map[string]bool)
    
    var dfs func(id string) bool
    dfs = func(id string) bool {
        visited[id] = true
        path[id] = true
        
        task, exists := allTasks[id]
        if !exists {
            return false
        }
        
        for _, depID := range task.DependsOn {
            if path[depID] {
                return true // cycle found
            }
            if !visited[depID] && dfs(depID) {
                return true
            }
        }
        
        path[id] = false
        return false
    }
    
    // Add the new task temporarily
    allTasks[taskID] = &Task{DependsOn: dependsOn}
    defer delete(allTasks, taskID)
    
    if dfs(taskID) {
        return fmt.Errorf("cycle detected involving task %s", taskID)
    }
    return nil
}
```

### 7.2 Task Execution Order

```go
// Topological sort respecting cross-project dependencies
func executionOrder(tasks []*Task) []*Task {
    // 1. Build adjacency graph
    // 2. Topological sort (Kahn's algorithm)
    // 3. Within same topological level, sort by sequence number
    // 4. Return ordered list
    
    // Tasks from different projects at the same level 
    // CAN run in parallel (different goroutines)
    // Tasks from the same project are always sequential
}
```

### 7.3 Unblock Dependents

```go
// After task completes, check if any blocked tasks can now run
func unblockDependents(completedTaskID string) {
    // 1. Find all tasks where depends_on contains completedTaskID
    // 2. For each dependent task:
    //    a. Load all its dependencies
    //    b. Check if ALL are status "done"
    //    c. If yes: change status from "blocked" to "queued"
    // 3. This may trigger cross-project unblocking
}
```

---

## 8. Testing Strategy

### 8.1 Test Types

| Type | What | How |
|------|------|-----|
| **Entity tests** | Entity creation, validation, relations | Real DB (yolotest) |
| **Action tests** | Business logic in actions | Real DB (yolotest) |
| **Service tests** | Orchestrator, dependency resolver, context builder | Real DB + mock claude CLI |
| **Job tests** | Workflow execution, timeout handling | Real DB + mock claude CLI |
| **CLI tests** | Command parsing, output formatting | Integration tests |
| **E2E tests** | Full workflow: PRD → tasks → execution → completion | Real DB + mock claude CLI |

### 8.2 Mock Claude CLI

For testing, Factory needs a mock `claude` CLI that:
- Accepts the same flags
- Returns valid JSON responses
- Simulates: success, failure, timeout, question
- Configurable per test case

```go
// test helper
func mockClaude(t *testing.T, response AgentResult) string {
    // Creates a temporary script that outputs the JSON response
    // Returns path to the script
    // Set CLAUDE_CLI_PATH env var to use this instead of real claude
}
```

### 8.3 Integration Test Flow

```go
func TestFullWorkflow(t *testing.T) {
    // 1. Register project
    // 2. Submit PRD
    // 3. Mock planner → returns task list
    // 4. Verify tasks created with correct dependencies
    // 5. Execute PRD
    // 6. Mock implementer → returns success
    // 7. Mock reviewer → returns pass
    // 8. Verify task status → done
    // 9. Verify PRD status → completed
    // 10. Verify backup written
}
```

---

## 9. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Bad PRD → bad tasks → wasted money | High | Medium | /grill-me before PRDs, human reviews task breakdown |
| Agent writes code that passes tests but is wrong | Medium | High | Two-stage review, acceptance criteria specificity |
| Claude API outage | Low | High | Exponential backoff, pause project, notify human |
| Budget runaway | Medium | Medium | Per-task + per-PRD + monthly caps, auto-stop |
| Context rot within long tasks | Medium | Medium | Steps are separate sessions, fresh context per step |
| Cross-project dependency deadlock | Low | High | Cycle detection, cascade failure on dep failure |
| Factory DB corruption | Low | High | YAML backup, daily snapshots, recovery command |
| Git merge conflicts | Low (sequential) | Medium | Worktrees, branch-per-task, clean merge |
| CLAUDE.md drift between tasks | Low | Medium | Snapshot CLAUDE.md at PRD start |
| Agent modifies wrong files | Medium | Medium | Review step catches scope creep, worktrees isolate |
| Sentinel creates too many tasks | Low | Low | Trust levels, suggestion queue, human approval |
| Token costs higher than estimated | Medium | Low | Budget caps, cost monitoring, advisor optimization |
