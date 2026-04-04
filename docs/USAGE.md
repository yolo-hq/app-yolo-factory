# YOLO Factory — Usage Guide

## Prerequisites

- Go 1.22+
- PostgreSQL (or SQLite for local dev)
- Redis (for job queue)
- Claude Code CLI (`claude`) installed and authenticated
- Git

## Setup

### 1. Database

```bash
# PostgreSQL
createdb yolo_factory
export DATABASE_URL="postgresql://postgres@localhost:5432/yolo_factory?sslmode=disable"

# Run migrations
yolo migrate up
```

### 2. Start Factory

```bash
# Start API server (port 9000) + MCP (port 3001)
yolo run api

# Start worker (processes jobs)
yolo run worker

# Start admin UI (port 3000)
yolo run admin
```

Or all at once:

```bash
yolo dev
```

---

## Quick Start

### Register a project

```bash
factory project add \
  --name plugin-webhooks \
  --repo git@github.com:yolo-hq/plugin-webhooks.git \
  --path /repos/plugin-webhooks \
  --branch main \
  --model sonnet
```

### Submit a PRD

From a file:
```bash
factory prd submit \
  --project plugin-webhooks \
  --file ./prd-webhook-retry.md
```

Or inline:
```bash
factory prd submit \
  --project plugin-webhooks \
  --title "Webhook retry with exponential backoff" \
  --body "Add retry logic for failed webhook deliveries..." \
  --criteria "Retries up to configured max" \
  --criteria "Exponential backoff between retries"
```

### Execute the PRD

```bash
# 1. Approve the PRD
factory prd approve <prd-id>

# 2. Execute (triggers planning → task creation → execution)
factory prd execute <prd-id>
```

If the project has `auto_start: true`, approving the PRD automatically triggers execution.

### Watch progress

```bash
factory status

# Output:
# Factory Status
# ──────────────
# Running:  1 task (plugin-webhooks:task-003)
# Queued:   3 tasks
# Done:     2 tasks
# Today:    $1.84 spent
```

---

## The Workflow

When you execute a PRD, Factory does this:

```
1. PLANNING
   Planner agent (Opus) reads your PRD + codebase
   → Creates ordered tasks with dependencies

2. PER TASK (sequential):
   a. Plan    — Opus reads code, outputs implementation plan
   b. Implement — Sonnet writes code via TDD (resumes plan session)
   c. Test    — go build + go test (shell, zero tokens)
   d. Lint    — factory lint checks (AST + grep, zero tokens)
   e. Audit   — Sonnet checks YOLO conventions
   f. Review  — Sonnet verifies acceptance criteria with evidence
   
   If any step fails → retry with error feedback
   After N retries → escalate model (Sonnet → Opus)
   After max retries → fail task, notify

3. COMPLETION
   Merge to target branch, push, unblock dependent tasks
   After all tasks → PRD alignment review
```

---

## CLI Reference

### Projects

```bash
factory project add --name X --repo X --path X [--branch main] [--model sonnet]
factory project list [--status active]
factory project get <id-or-name>
factory project update <id> [--auto-merge true] [--budget-monthly 300]
factory project pause <id>
factory project resume <id>
factory project archive <id>
```

### PRDs

```bash
factory prd submit --project X --file ./prd.md
factory prd submit --project X --title X --body X --criteria "X"
factory prd list [--project X] [--status in_progress]
factory prd get <id>
factory prd approve <id>
factory prd execute <id>
```

### Tasks

```bash
factory task list [--prd X] [--project X] [--status queued]
factory task get <id>
factory task cancel <id>
factory task retry <id> [--model opus]
```

### Status & Cost

```bash
factory status
factory cost [--period month] [--project X]
```

### Questions

When an agent gets confused, it creates a question. Answer it to unblock:

```bash
factory questions list [--status open]
factory questions answer <id> "Use context.Context for cancellation"
```

### Suggestions

Sentinel and Advisor create suggestions. Review and promote:

```bash
factory suggestions list [--project X] [--category optimization]
factory suggestions approve <id> [--prd X]
factory suggestions reject <id> --reason "Not needed"
```

### Insights

Process Advisor generates insights from execution history:

```bash
factory insight list [--category cost_optimization]
factory insight acknowledge <id>
factory insight apply <id>
factory insight dismiss <id> --reason "Not applicable"
```

### Background Agents

```bash
# Run health checks manually
factory sentinel run [--project X]

# Run optimization analysis manually
factory advisor run [--project X]

# Manual backup
factory backup

# Recover from backup
factory recover --from /path/to/factory-state
```

---

## Project Configuration

```bash
factory project add --name my-project --repo git@github.com:org/repo.git --path /repos/repo
factory project update <id> \
  --auto-merge true \          # merge on success (default: true)
  --auto-start false \         # auto-execute on PRD approval (default: false)
  --budget-monthly 200 \       # monthly cost cap in USD
  --model sonnet \             # default model
  --max-retries 3              # retries before failing
```

### Key settings

| Setting | Default | Description |
|---------|---------|-------------|
| `default_model` | sonnet | Model for implementation |
| `escalation_model` | opus | Model after N failed retries |
| `escalation_after_retries` | 2 | When to escalate |
| `budget_per_task_usd` | 2.00 | Max cost per task |
| `budget_monthly_usd` | 200.00 | Monthly cap |
| `budget_warning_at` | 0.8 | Warn at 80% |
| `auto_merge` | true | Merge on success |
| `auto_start` | false | Auto-execute on PRD approval |
| `max_retries` | 3 | Max retries per task |
| `timeout_secs` | 600 | Task timeout (10 min) |
| `use_worktrees` | false | Git worktrees (off for linked modules) |
| `push_failed_branches` | false | Push branches on failure |

---

## Writing PRDs for Factory

A good PRD has:

```markdown
## Title
Short, descriptive

## Body
What to build and why. Include:
- Context the agent needs
- Constraints and boundaries
- What NOT to do

## Acceptance Criteria
Specific, testable conditions:
- "RetryPolicy interface exists in core/retry/"
- "Exponential backoff with jitter in default implementation"
- "Integration test covers retry with failing endpoint"

## Design Decisions (optional)
- "RetryPolicy in core, not plugin — reusable"
- "Exponential backoff — prevents thundering herd"
```

**Tips:**
- More acceptance criteria = higher success rate
- Criteria should be verifiable by reading code (the review agent checks)
- Include file paths if you know them
- Include code patterns to follow
- Keep each task small — Factory breaks PRDs into tasks, smaller tasks succeed more

---

## Quality Gates

Every task passes through 6 gates:

| Gate | Type | Cost | What it checks |
|------|------|------|---------------|
| **Plan** | Agent (Opus) | ~$0.40 | Creates implementation plan |
| **Implement** | Agent (Sonnet) | ~$0.30 | Writes code via TDD |
| **Test** | Program | $0 | `go build` + `go test` |
| **Lint** | Program | $0 | Swallowed errors, shell injection, status literals, stubs, duplicates |
| **Audit** | Agent (Sonnet) | ~$0.08 | YOLO conventions |
| **Review** | Agent (Sonnet) | ~$0.15 | Acceptance criteria with file:line evidence |

**Cost per task: ~$0.93**

Program gates catch 80% of issues at zero token cost. Agent gates handle judgment calls.

---

## Background Systems

### Sentinel (auto-healing)
Runs daily. Checks:
- Build health (`go build`)
- Test health (`go test`)
- Security (`govulncheck`)
- Convention drift (`yolo audit`)
- Orphaned runs (stuck tasks)

High-trust issues (broken build) → auto-creates fix tasks.
Medium-trust issues (conventions) → creates suggestions for review.

### Advisor (optimization)
Runs weekly. Suggests:
- Pattern extraction ("this retry logic could be shared")
- Code quality trends
- Refactoring opportunities

### Process Advisor (self-improvement)
Runs weekly after 20+ completed tasks. Analyzes:
- Retry rates by project/model
- Cost per step type
- Model effectiveness (Opus vs Sonnet success rates)
- Gate effectiveness (which gates catch real issues)
- Spec quality correlation (more criteria = more success?)

Creates Insight entities with data-backed recommendations.

---

## Backup & Recovery

Factory auto-backs up state to a git repository:

```
factory-state/
├── projects/plugin-webhooks.yml
├── prds/prd-001.yml
├── tasks/task-001.yml     # includes nested runs/steps/reviews
├── questions/
├── suggestions/
└── snapshots/2026-04-04.yml
```

- Every state change → YAML file + git commit
- Daily → full snapshot
- Recovery: `factory recover --from ./factory-state`

---

## MCP Access

Factory exposes MCP tools on port 3001. From any Claude Code session:

```
# In your Claude Code MCP config, add Factory:
{
  "factory": {
    "url": "http://localhost:3001"
  }
}
```

Then in Claude Code:
```
> What's Factory working on?
(Claude calls factory_status tool)

> Submit this PRD to Factory
(Claude calls factory_submit_prd tool)

> Are there any open questions?
(Claude calls factory_list_questions tool)
```

---

## Troubleshooting

### Task keeps failing
```bash
factory task get <id>          # check error message
factory task retry <id> --model opus  # retry with stronger model
```

### Budget exceeded
```bash
factory cost --project X       # check spending
factory project update <id> --budget-monthly 500  # increase limit
```

### Agent asked a question
```bash
factory questions list --status open
factory questions answer <id> "Your answer here"
```

### Factory stuck
```bash
factory status                 # check running tasks
factory sentinel run --all     # check for orphaned runs
```
