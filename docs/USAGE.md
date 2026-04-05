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
createdb yolo_factory
export DATABASE_URL="postgresql://postgres@localhost:5432/yolo_factory?sslmode=disable"
yolo migrate up
```

### 2. Start Factory

```bash
yolo dev
```

This starts the API server (port 9000), MCP server (port 3001), worker, and admin UI (port 3000).

### 3. Set up the alias

```bash
alias yf="yolo run factory"
```

Add this to your `.zshrc` or `.bashrc`.

### 4. Connecting Claude Code (MCP)

Add Factory as an MCP server so Claude Code can interact with it directly:

```bash
claude mcp add factory --transport sse --url http://localhost:3001/mcp
```

Then in any Claude Code session you can ask "What's Factory working on?" or use `/factory:submit` to send PRDs.

---

## Register Projects

### Scan a directory

```bash
# Preview what would be registered
yf project:scan --dir ~/projects/yolo-hq --dry-run

# Register all repos
yf project:scan --dir ~/projects/yolo-hq
```

This finds all directories with both `.git/` and `go.mod`, and registers them as Factory projects.

### Add a single project

```bash
yf project:add \
  --name plugin-webhooks \
  --repo git@github.com:yolo-hq/plugin-webhooks.git \
  --path ~/projects/yolo-hq/plugin-webhooks
```

### Configure a project

```bash
yf project:update <id> --model sonnet --budget 300 --retries 3
```

---

## How PRDs Become Tasks

```
You write a PRD
  │
  ▼
yf prd:submit → PRD created (draft)
  │
  ▼
yf prd:execute → Planner agent (Opus) reads PRD + codebase
  │               breaks it into ordered tasks
  ▼
Tasks execute sequentially:
  Plan → Implement → Test → Lint → Audit → Review
  │
  ▼
Each passing task merges to target branch
  │
  ▼
All tasks done → PRD complete
```

---

## Submit a PRD

From a file:
```bash
yf prd:submit --project plugin-webhooks --file ./prd.md --title "Webhook retry"
```

Inline:
```bash
yf prd:submit \
  --project plugin-webhooks \
  --title "Webhook retry with exponential backoff" \
  --body "Add retry logic for failed webhook deliveries..." \
  --criteria "Retries up to configured max"
```

From Claude Code (after `/grill-me`):
```
/factory:submit
```

### Writing good PRDs

- **Acceptance criteria are key.** More criteria = higher success rate. Minimum 3.
- **Criteria must be code-verifiable.** The review agent checks with file:line evidence.
- **Include file paths** if you know them.
- **Include code patterns** to follow.
- **Keep it concise.** Agents read it — long specs waste tokens.

---

## Execute and Monitor

```bash
# Execute a PRD (triggers planning → tasks → execution)
yf prd:execute <prd-id>

# Watch progress
yf status

# View tasks for a PRD
yf task:list --prd <prd-id>

# View a specific task
yf task:get <task-id>

# See what Factory changed
yf prd:diff --id <prd-id>
```

---

## Interact

### Answer questions

When an agent gets confused, it creates a question. Answer it to unblock:

```bash
yf questions:list
yf questions:answer <id> "Use context.Context for cancellation"
```

### Review suggestions

Sentinel and Advisor create suggestions. Review and promote:

```bash
yf suggestions:list
yf suggestions:approve <id>
yf suggestions:reject <id> --reason "Not needed"
```

### Process insights

After 20+ completed tasks, the Process Advisor generates data-backed insights:

```bash
yf insight:list
yf insight:acknowledge <id>
yf insight:apply <id>
```

---

## Control Execution

```bash
# Pause/resume a project
yf project:pause <id>
yf project:resume <id>

# Cancel a task
yf task:cancel <id>

# Retry a failed task (optionally with a stronger model)
yf task:retry <id> --model opus

# Cost report
yf cost --period month --project plugin-webhooks
```

---

## Background Systems

### Sentinel (auto-healing)
Runs daily. Checks build health, test health, security, convention drift, orphaned runs. High-trust issues get auto-fix tasks. Medium-trust issues become suggestions.

```bash
yf sentinel:run  # run manually
```

### Advisor (optimization)
Runs weekly. Suggests pattern extraction, refactoring, code quality improvements.

```bash
yf advisor:run  # run manually
```

---

## CLI Reference

```bash
# Projects
yf project:scan --dir <path> [--dry-run] [--branch main] [--model sonnet]
yf project:add --name X --repo X --path X [--branch main] [--model sonnet]
yf project:list
yf project:get <id-or-name>
yf project:update <id> [--model X] [--budget X] [--retries X]
yf project:pause <id>
yf project:resume <id>
yf project:archive <id>

# PRDs
yf prd:submit --project X --file prd.md [--title X]
yf prd:submit --project X --title X --body X --criteria "X"
yf prd:list [--project X] [--status in_progress]
yf prd:get <id>
yf prd:approve <id>
yf prd:execute <id>
yf prd:diff --id <id>

# Tasks
yf task:list [--prd X] [--project X] [--status queued]
yf task:get <id>
yf task:cancel <id>
yf task:retry <id> [--model opus]

# Status & Cost
yf status
yf cost [--period month] [--project X]

# Questions
yf questions:list [--status open]
yf questions:answer <id> "answer"

# Suggestions
yf suggestions:list [--project X]
yf suggestions:approve <id>
yf suggestions:reject <id> --reason "..."

# Insights
yf insight:list [--category cost_optimization]
yf insight:acknowledge <id>
yf insight:apply <id>
yf insight:dismiss <id> --reason "..."

# Background
yf sentinel:run [--project X]
yf advisor:run [--project X]
yf backup
yf recover --from /path/to/factory-state
```

---

## Project Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `default_model` | sonnet | Model for implementation |
| `escalation_model` | opus | Model after N failed retries |
| `escalation_after_retries` | 2 | When to escalate |
| `budget_per_task_usd` | 2.00 | Max cost per task |
| `budget_monthly_usd` | 200.00 | Monthly cap |
| `auto_merge` | true | Merge on success |
| `auto_start` | false | Auto-execute on PRD approval |
| `max_retries` | 3 | Max retries per task |
| `timeout_secs` | 600 | Task timeout (10 min) |
