# YOLO Factory — Software Requirements Specification

**Version:** 1.0.0
**Date:** 2026-04-04
**Status:** Draft

---

## 1. Introduction

### 1.1 Purpose

YOLO Factory is an autonomous software development engine built on the YOLO framework. It takes a PRD (Product Requirements Document) as input and delivers a completed project — planned, implemented, tested, reviewed, and merged — with minimal human intervention.

### 1.2 Scope

Factory manages the full lifecycle of software development tasks across multiple YOLO projects:

- YOLO core framework
- YOLO plugins (plugin-webhooks, plugin-rabbitmq, etc.)
- YOLO applications (app-libromi, etc.)

It orchestrates Claude Code CLI agents to plan, implement, test, review, and merge code changes. It handles cross-project dependencies, versioned branches, cost tracking, self-healing, and intelligent suggestions.

### 1.3 Definitions

| Term | Definition |
|------|-----------|
| **PRD** | Product Requirements Document — structured spec with acceptance criteria |
| **Task** | One unit of work producing one commit in one repo on one branch |
| **Run** | One execution attempt of a task (a task may have multiple runs on retry) |
| **Step** | One phase within a run (plan, implement, test, audit, review) |
| **Agent** | A Claude Code CLI session spawned by Factory to do work |
| **Skill** | A reusable prompt/workflow that an agent follows (e.g., TDD, audit) |
| **Sentinel** | Background agent that watches for problems and auto-creates fix tasks |
| **Advisor** | Background agent that suggests optimizations, refactoring, and new ideas |
| **Worktree** | Git worktree — isolated working directory for a branch |
| **Factory State** | Backup git repo containing YAML snapshots of all Factory data |

### 1.4 Design Principles

1. **Sequential per project, parallel across projects** — one task at a time per project; different projects can run simultaneously since they are separate repos
2. **Spec-driven** — human writes spec (PRD), agents write code
3. **Fresh context per step** — each agent step gets a clean context window to prevent context rot
4. **Implementation != Review** — the agent that writes code never reviews its own work
5. **Convention-enforced** — every task passes `yolo audit` and build/test gates
6. **Observable** — every agent session, cost, token count, and decision is tracked
7. **Recoverable** — DB primary, YAML backup; failed tasks leave no mess on main branch
8. **Agent-agnostic execution** — Factory orchestrates via CLI subprocess; the agent doesn't know Factory exists

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
Human
  │
  ├── /grill-me → refine idea
  ├── /write-prd → create PRD
  └── Factory CLI/UI → submit PRD, manage projects, review output
        │
        ▼
┌─────────────────────────────────────────────┐
│              YOLO Factory                    │
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐ │
│  │ API      │  │ Worker   │  │ CLI       │ │
│  │ (HTTP)   │  │ (Jobs)   │  │ (Commands)│ │
│  │ (MCP)    │  │          │  │           │ │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘ │
│       │              │              │        │
│       └──────┬───────┘──────────────┘        │
│              │                               │
│       ┌──────┴──────┐                        │
│       │  PostgreSQL  │                        │
│       └──────┬──────┘                        │
│              │                               │
│       ┌──────┴──────┐                        │
│       │ YAML Backup │ → factory-state repo   │
│       └─────────────┘                        │
│                                              │
│  ┌──────────────────────────────────────┐   │
│  │         Agent Orchestrator            │   │
│  │                                       │   │
│  │  Go CLI Wrapper (pkg/claude/)         │   │
│  │    └── spawns claude CLI subprocess   │   │
│  │    └── parses JSON output             │   │
│  │    └── manages sessions               │   │
│  └──────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
        │
        ▼
┌─────────────────┐
│  Claude Code CLI │ (installed separately)
│  --bare mode     │
│  --output json   │
└─────────────────┘
        │
        ▼
┌─────────────────┐
│  Target Repos    │
│  (yolo-core,     │
│   plugins, apps) │
└─────────────────┘
```

### 2.2 Components

| Component | Technology | Purpose |
|-----------|-----------|---------|
| API Server | YOLO HTTP + MCP listeners | External access, admin UI, MCP tools |
| Worker | YOLO asynq (Redis-backed) | Background job execution |
| CLI | YOLO cobra commands | Local management, manual operations |
| Database | PostgreSQL (SQLite for local dev) | Primary data store |
| Agent Wrapper | Go package (`pkg/claude/`) | Spawns and manages Claude Code CLI |
| Backup | Git repo with YAML files | Disaster recovery, audit trail |
| Admin UI | YOLO admin (@yolo-hq/admin) | Dashboard, project management |
| Notification | plugin-notifications | Configurable alerts (slack, email, webhook) |

### 2.3 Integration Points

| System | Integration | Direction |
|--------|------------|-----------|
| Claude Code CLI | Subprocess via `exec.Command` | Factory → CLI |
| GitHub | Git push, PR creation via `gh` CLI | Factory → GitHub |
| YOLO Admin | Entity endpoints, UI YAML | Factory → Admin |
| plugin-notifications | YOLO event system | Factory → Plugin |
| MCP Clients | Factory as MCP server | External → Factory |
| factory-state repo | Git push YAML backups | Factory → Git |

---

## 3. Entity Specifications

### 3.1 Project

Represents a git repository that Factory manages.

```yaml
entity: Project
table: factory_projects

fields:
  id:                    { type: ulid, primary: true }
  name:                  { type: string, unique: true, required: true }
  status:                { type: string, enum: [active, paused, archived], default: active }

  # Repository
  repo_url:              { type: string, required: true }
  local_path:            { type: string, required: true }
  use_worktrees:         { type: bool, default: false }

  # Branches
  default_branch:        { type: string, default: "main" }
  maintenance_branches:  { type: json, default: "[]" }  # ["v1.x", "v2.x"]

  # Agent defaults
  default_model:         { type: string, enum: [sonnet, opus, haiku], default: sonnet }
  escalation_model:      { type: string, enum: [sonnet, opus], default: opus }
  escalation_after_retries: { type: int, default: 2 }

  # Budget
  budget_per_task_usd:   { type: float64, default: 2.0 }
  budget_per_prd_usd:    { type: float64, default: 20.0 }
  budget_monthly_usd:    { type: float64, default: 200.0 }
  budget_warning_at:     { type: float64, default: 0.8 }  # notify at 80%
  spent_this_month_usd:  { type: float64, default: 0.0 }

  # Execution
  max_retries:           { type: int, default: 3 }
  timeout_secs:          { type: int, default: 600 }
  auto_merge:            { type: bool, default: true }
  auto_start:            { type: bool, default: false }
  push_failed_branches:  { type: bool, default: false }

  # Environment
  setup_commands:        { type: json, default: "[]" }  # ["go mod download"]
  test_commands:         { type: json, default: '["go build ./...", "go test ./..."]' }

  # Timestamps
  created_at:            { type: timestamp, auto: true }
  updated_at:            { type: timestamp, auto: true }

relations:
  prds:        { type: has_many, entity: PRD }
  tasks:       { type: has_many, entity: Task }
  suggestions: { type: has_many, entity: Suggestion }
```

### 3.2 PRD

Product Requirements Document — the input spec for Factory.

```yaml
entity: PRD
table: factory_prds

fields:
  id:                    { type: ulid, primary: true }
  project_id:            { type: ulid, required: true }
  title:                 { type: string, required: true }
  status:                { type: string, enum: [draft, approved, planning, in_progress, completed, failed], default: draft }

  # Source
  source:                { type: string, enum: [manual, grill_me, factory_generated, imported], default: manual }
  created_by:            { type: string, enum: [human, advisor, sentinel], default: human }

  # Spec
  body:                  { type: text, required: true }
  acceptance_criteria:   { type: json, required: true }  # [{id, description, verification}]
  design_decisions:      { type: json, default: "[]" }   # ["decision 1", "decision 2"]

  # Tracking
  total_tasks:           { type: int, default: 0 }
  completed_tasks:       { type: int, default: 0 }
  failed_tasks:          { type: int, default: 0 }
  total_cost_usd:        { type: float64, default: 0.0 }

  # Timestamps
  approved_at:           { type: timestamp, nullable: true }
  completed_at:          { type: timestamp, nullable: true }
  created_at:            { type: timestamp, auto: true }
  updated_at:            { type: timestamp, auto: true }

relations:
  project: { type: belongs_to, entity: Project }
  tasks:   { type: has_many, entity: Task }
```

### 3.3 Task

One unit of work = one commit in one repo on one branch.

```yaml
entity: Task
table: factory_tasks

fields:
  id:                    { type: ulid, primary: true }
  prd_id:                { type: ulid, required: true }
  project_id:            { type: ulid, required: true }
  title:                 { type: string, required: true }
  status:                { type: string, enum: [queued, blocked, running, reviewing, done, failed, cancelled], default: queued }

  # Spec
  spec:                  { type: text, required: true }
  acceptance_criteria:   { type: json, required: true }  # [{id, description}]

  # Execution config
  branch:                { type: string, required: true }  # target branch (e.g., "main", "v1.x")
  model:                 { type: string, nullable: true }   # override project default
  sequence:              { type: int, required: true }      # order within PRD

  # Dependencies (cross-project supported)
  depends_on:            { type: json, default: "[]" }  # ["task-id-1", "other-project:task-id-2"]

  # Tracking
  run_count:             { type: int, default: 0 }
  max_retries:           { type: int, default: 3 }
  cost_usd:              { type: float64, default: 0.0 }

  # Filled on completion
  summary:               { type: text, nullable: true }  # what was done (context for downstream tasks)
  commit_hash:           { type: string, nullable: true }

  # Timestamps
  started_at:            { type: timestamp, nullable: true }
  completed_at:          { type: timestamp, nullable: true }
  created_at:            { type: timestamp, auto: true }
  updated_at:            { type: timestamp, auto: true }

relations:
  prd:        { type: belongs_to, entity: PRD }
  project:    { type: belongs_to, entity: Project }
  runs:       { type: has_many, entity: Run }
  questions:  { type: has_many, entity: Question }
  parent:     { type: belongs_to, entity: Task, field: depends_on }
```

### 3.4 Run

One execution attempt of a task. A task may have multiple runs on retry.

```yaml
entity: Run
table: factory_runs

fields:
  id:                    { type: ulid, primary: true }
  task_id:               { type: ulid, required: true }
  agent_type:            { type: string, enum: [planner, implementer, reviewer, auditor, sentinel, advisor], required: true }
  status:                { type: string, enum: [running, completed, failed, cancelled], default: running }

  # Agent config
  model:                 { type: string, required: true }
  session_id:            { type: string, nullable: true }  # Claude Code session UUID
  session_name:          { type: string, nullable: true }  # factory:{task}:{phase}
  escalated_model:       { type: string, nullable: true }  # if auto-escalated after retry

  # Metrics
  cost_usd:              { type: float64, default: 0.0 }
  tokens_in:             { type: int, default: 0 }
  tokens_out:            { type: int, default: 0 }
  duration_ms:           { type: int, default: 0 }
  num_turns:             { type: int, default: 0 }

  # Git
  commit_hash:           { type: string, nullable: true }
  branch_name:           { type: string, nullable: true }
  files_changed:         { type: json, default: "[]" }  # captured before commit/discard

  # Result
  result:                { type: text, nullable: true }  # agent's final response
  error:                 { type: text, nullable: true }

  # Timestamps
  started_at:            { type: timestamp, auto: true }
  completed_at:          { type: timestamp, nullable: true }

relations:
  task:    { type: belongs_to, entity: Task }
  steps:   { type: has_many, entity: Step }
  review:  { type: has_one, entity: Review }
```

### 3.5 Step

One workflow phase within a run.

```yaml
entity: Step
table: factory_steps

fields:
  id:                    { type: ulid, primary: true }
  run_id:                { type: ulid, required: true }
  phase:                 { type: string, enum: [plan, implement, test, audit, review], required: true }
  skill:                 { type: string, required: true }  # skill name used (plan-task, tdd, audit, review-task)
  status:                { type: string, enum: [running, completed, failed, skipped], default: running }

  # Agent
  model:                 { type: string, required: true }
  session_id:            { type: string, nullable: true }

  # Metrics
  cost_usd:              { type: float64, default: 0.0 }
  tokens_in:             { type: int, default: 0 }
  tokens_out:            { type: int, default: 0 }
  duration_ms:           { type: int, default: 0 }

  # Context
  input_summary:         { type: text, nullable: true }
  output_summary:        { type: text, nullable: true }

  # Timestamps
  started_at:            { type: timestamp, auto: true }
  completed_at:          { type: timestamp, nullable: true }

relations:
  run: { type: belongs_to, entity: Run }
```

### 3.6 Review

Reviewer agent's verdict on a run.

```yaml
entity: Review
table: factory_reviews

fields:
  id:                    { type: ulid, primary: true }
  run_id:                { type: ulid, required: true }
  task_id:               { type: ulid, required: true }

  # Reviewer
  session_id:            { type: string, nullable: true }
  model:                 { type: string, required: true }

  # Verdict
  verdict:               { type: string, enum: [pass, fail], required: true }
  reasons:               { type: json, default: "[]" }             # if failed, why
  anti_patterns:         { type: json, default: "[]" }             # anti-patterns found
  criteria_results:      { type: json, required: true }            # [{criteria_id, passed, comment}]
  suggestions:           { type: json, default: "[]" }             # optional non-blocking improvements

  # Metrics
  cost_usd:              { type: float64, default: 0.0 }

  # Timestamps
  created_at:            { type: timestamp, auto: true }

relations:
  run:  { type: belongs_to, entity: Run }
  task: { type: belongs_to, entity: Task }
```

### 3.7 Question

Agent question raised during execution + answer chain.

```yaml
entity: Question
table: factory_questions

fields:
  id:                    { type: ulid, primary: true }
  task_id:               { type: ulid, required: true }
  run_id:                { type: ulid, required: true }

  # Question
  body:                  { type: text, required: true }
  context:               { type: text, nullable: true }  # what agent was doing when question arose
  confidence:            { type: string, enum: [low, medium], required: true }
    # low = pause execution, wait for answer
    # medium = agent proceeds with best guess, flag for review

  # Answer
  status:                { type: string, enum: [open, answered, auto_resolved], default: open }
  answer:                { type: text, nullable: true }
  answered_by:           { type: string, enum: [human, planner, auto], nullable: true }
  answer_session_id:     { type: string, nullable: true }

  # Timestamps
  created_at:            { type: timestamp, auto: true }
  answered_at:           { type: timestamp, nullable: true }

relations:
  task: { type: belongs_to, entity: Task }
  run:  { type: belongs_to, entity: Run }
```

### 3.8 Suggestion

Sentinel/Advisor recommendations.

```yaml
entity: Suggestion
table: factory_suggestions

fields:
  id:                    { type: ulid, primary: true }
  project_id:            { type: ulid, required: true }

  # Source
  source:                { type: string, enum: [sentinel, advisor], required: true }
  category:              { type: string, enum: [optimization, refactoring, tech_debt, security, new_feature, pattern_extraction, bug_fix], required: true }

  # Content
  title:                 { type: string, required: true }
  body:                  { type: text, required: true }
  priority:              { type: string, enum: [low, medium, high, critical], default: medium }

  # Lifecycle
  status:                { type: string, enum: [pending, approved, rejected, converted], default: pending }
  converted_task_id:     { type: ulid, nullable: true }

  # Timestamps
  created_at:            { type: timestamp, auto: true }
  updated_at:            { type: timestamp, auto: true }

relations:
  project:        { type: belongs_to, entity: Project }
  converted_task: { type: belongs_to, entity: Task, nullable: true }
```

### 3.9 Entity Relationships

```
Project (1) ──── (many) PRD
Project (1) ──── (many) Task
Project (1) ──── (many) Suggestion

PRD     (1) ──── (many) Task

Task    (1) ──── (many) Run
Task    (1) ──── (many) Question
Task    (many) ── (many) Task  (dependencies via depends_on JSON)

Run     (1) ──── (many) Step
Run     (1) ──── (0..1) Review
Run     (1) ──── (many) Question

Suggestion (0..1) ── (1) Task  (if converted)
```

---

## 4. Agent System

### 4.1 Agent Types

| Agent | Purpose | Model | Tools | Bare Mode | Budget |
|-------|---------|-------|-------|-----------|--------|
| **Planner** | Break PRD into tasks, create implementation plans | Opus | Read, Glob, Grep | Yes | $1.00 |
| **Implementer** | Write code following TDD | Sonnet | Read, Edit, Write, Bash, Glob, Grep | Yes | $2.00 |
| **Reviewer** | Check quality vs acceptance criteria | Sonnet | Read, Glob, Grep | Yes | $0.50 |
| **Auditor** | Run `yolo audit`, convention compliance | Sonnet | Read, Bash, Glob, Grep | Yes | $0.30 |
| **Sentinel** | Watch for breaks, auto-create fix tasks | Haiku | Read, Bash, Glob, Grep | No | $0.20 |
| **Advisor** | Suggest optimizations, ideas, refactoring | Sonnet | Read, Glob, Grep | No | $0.50 |

### 4.2 Agent Configuration

Stored in `app.yml` under `agent_profiles`:

```yaml
agent_profiles:
  planner:
    model: opus
    allowed_tools: ["Read", "Glob", "Grep"]
    bare: true
    budget_usd: 1.0
    permission_mode: "auto"
    effort: "high"

  implementer:
    model: sonnet
    allowed_tools: ["Read", "Edit", "Write", "Bash", "Glob", "Grep"]
    bare: true
    budget_usd: 2.0
    permission_mode: "auto"
    effort: "high"

  reviewer:
    model: sonnet
    allowed_tools: ["Read", "Glob", "Grep"]
    bare: true
    budget_usd: 0.5
    permission_mode: "plan"
    effort: "medium"

  auditor:
    model: sonnet
    allowed_tools: ["Read", "Bash", "Glob", "Grep"]
    bare: true
    budget_usd: 0.3
    permission_mode: "auto"
    effort: "medium"

  sentinel:
    model: haiku
    allowed_tools: ["Read", "Bash", "Glob", "Grep"]
    bare: false
    budget_usd: 0.2
    permission_mode: "auto"
    effort: "low"

  advisor:
    model: sonnet
    allowed_tools: ["Read", "Glob", "Grep"]
    bare: false
    budget_usd: 0.5
    permission_mode: "plan"
    effort: "medium"
```

### 4.3 Model Escalation

When a task fails and retries:

```
Attempt 1: Use configured model (e.g., Sonnet)
Attempt 2: Use configured model (e.g., Sonnet)
Attempt 3: Escalate to project.escalation_model (e.g., Opus)
```

Escalation threshold is configurable per project via `escalation_after_retries`.

### 4.4 Go CLI Wrapper

Package: `pkg/claude/`

```go
// pkg/claude/agent.go

type AgentConfig struct {
    Model          string
    AllowedTools   []string
    BudgetUSD      float64
    PermissionMode string   // "auto", "plan"
    Bare           bool
    Effort         string   // "low", "medium", "high", "max"
    CWD            string
    SystemPrompt   string   // appended to default
    SessionName    string   // --name flag
    ResumeSession  string   // --resume flag
    ForkSession    bool     // --fork-session flag
    JSONSchema     string   // --json-schema for structured output
    Env            []string // extra env vars
}

type AgentResult struct {
    Result           string   `json:"result"`
    SessionID        string   `json:"session_id"`
    CostUSD          float64  `json:"total_cost_usd"`
    InputTokens      int      `json:"input_tokens"`
    OutputTokens     int      `json:"output_tokens"`
    DurationMS       int      `json:"duration_ms"`
    NumTurns         int      `json:"num_turns"`
    IsError          bool     `json:"is_error"`
    StopReason       string   `json:"stop_reason"`
    StructuredOutput any      `json:"structured_output"`
}

// Run executes a one-shot agent session
func Run(ctx context.Context, config AgentConfig, prompt string) (*AgentResult, error)

// Resume continues an existing session with a new prompt
func Resume(ctx context.Context, sessionID string, config AgentConfig, prompt string) (*AgentResult, error)

// Fork creates a new session branched from an existing one
func Fork(ctx context.Context, sessionID string, config AgentConfig, prompt string) (*AgentResult, error)

// Stream executes and streams results via callback
func Stream(ctx context.Context, config AgentConfig, prompt string, cb func(event StreamEvent)) (*AgentResult, error)

// ListSessions lists Claude Code sessions matching a filter
func ListSessions(cwd string, filter string) ([]SessionInfo, error)
```

CLI command constructed internally:

```bash
claude -p \
  --bare \
  --output-format json \
  --model sonnet \
  --allowedTools "Read,Edit,Write,Bash,Glob,Grep" \
  --permission-mode auto \
  --max-budget-usd 2.0 \
  --effort high \
  --name "factory:task-001:implement" \
  --append-system-prompt "..." \
  "Implement webhook retry logic per this spec: ..."
```

### 4.5 Question Escalation Chain

When an agent raises a question during execution:

```
Step 1: Agent outputs question in structured result
        (detected via JSON schema or result parsing)

Step 2: Check auto-resolve
        Search CLAUDE.md, docs, previous task summaries
        If answer found → auto-resolve, resume agent

Step 3: Ask Planner agent (Opus)
        Spawn Planner with question + full context
        Planner returns answer
        Resume Implementer with --resume using answer

Step 4: Ask human (only if confidence = "low" AND Planner unsure)
        Notify human with question + what agents tried
        Pause task, wait for human response
        Resume Implementer with --resume using human answer
```

---

## 5. Workflow Specification

### 5.1 End-to-End Workflow

```
┌──────────────────────────────────────────────────────┐
│                    HUMAN INPUT                        │
│                                                       │
│  Option A: /grill-me → /write-prd → submit PRD       │
│  Option B: Write PRD manually → submit                │
│  Option C: One-liner → Factory generates PRD          │
│  Option D: Import from GitHub issue                   │
└───────────────────────┬──────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────┐
│                  PHASE 1: PLANNING                    │
│                                                       │
│  1. PRD status → "planning"                           │
│  2. Spawn Planner agent (Opus)                        │
│     Input: PRD body + acceptance criteria              │
│            + project CLAUDE.md + codebase exploration  │
│     Output: Ordered task list (structured JSON)        │
│  3. Create Task entities in DB                         │
│  4. Set dependencies between tasks                     │
│  5. Detect cycles (DFS) — reject if cycles found       │
│  6. PRD status → "approved" (if auto_start)            │
│     or wait for human approval                         │
│  7. Backup: write PRD + tasks to factory-state repo    │
└───────────────────────┬──────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────┐
│            PHASE 2: EXECUTION (per task)              │
│                                                       │
│  For each task in dependency order:                   │
│                                                       │
│  ┌─ PRE-CHECKS ────────────────────────────────┐     │
│  │ a. Verify all dependencies are "done"        │     │
│  │ b. Check project budget not exceeded          │     │
│  │ c. Update main branch: git pull               │     │
│  │ d. Create branch: task-{task_id}              │     │
│  │ e. Create worktree (if project.use_worktrees) │     │
│  │ f. Task status → "running"                    │     │
│  │ g. Create Run entity                          │     │
│  └──────────────────────────────────────────────┘     │
│                                                       │
│  ┌─ STEP 1: PLAN ──────────────────────────────┐     │
│  │ Agent: Planner (Opus, read-only)             │     │
│  │ Input: task spec + previous task summaries    │     │
│  │        + codebase state                       │     │
│  │ Output: implementation plan with file list    │     │
│  │ Session: new session (--name factory:{t}:plan)│     │
│  └──────────────────────────────────────────────┘     │
│                                                       │
│  ┌─ STEP 2: IMPLEMENT ────────────────────────┐      │
│  │ Agent: Implementer (Sonnet, full tools)      │     │
│  │ Input: task spec + plan output + TDD skill   │     │
│  │ Skill: factory:implement (headless TDD)      │     │
│  │ Session: --resume plan session (continues)    │     │
│  │ If question raised → escalation chain         │     │
│  └──────────────────────────────────────────────┘     │
│                                                       │
│  ┌─ STEP 3: TEST ─────────────────────────────┐      │
│  │ Run project.test_commands:                   │     │
│  │   go build ./...                             │     │
│  │   go test ./...                              │     │
│  │ If fail → retry Step 2 with error output     │     │
│  │ Not an agent step — direct shell execution   │     │
│  └──────────────────────────────────────────────┘     │
│                                                       │
│  ┌─ STEP 4: AUDIT ────────────────────────────┐      │
│  │ Agent: Auditor (Sonnet, read + bash)         │     │
│  │ Input: changed files + CLAUDE.md conventions  │     │
│  │ Skill: factory:audit (headless)              │     │
│  │ Session: new session                          │     │
│  │ Output: pass/fail + violations list           │     │
│  │ If fail → retry Step 2 with violations       │     │
│  └──────────────────────────────────────────────┘     │
│                                                       │
│  ┌─ STEP 5: REVIEW ───────────────────────────┐      │
│  │ Agent: Reviewer (Sonnet, read-only)          │     │
│  │ Input: git diff + acceptance criteria         │     │
│  │        + anti-pattern checklist               │     │
│  │ Skill: factory:review (headless)             │     │
│  │ Session: new session (fresh, isolated)        │     │
│  │ Output: Review entity (pass/fail + reasons)   │     │
│  │ If fail → retry Step 2 with review feedback   │     │
│  └──────────────────────────────────────────────┘     │
│                                                       │
│  ┌─ POST-TASK ─────────────────────────────────┐     │
│  │ If all steps pass:                            │     │
│  │   a. Git commit with task summary             │     │
│  │   b. Merge branch to target (if auto_merge)   │     │
│  │   c. Push to remote                           │     │
│  │   d. Task status → "done"                     │     │
│  │   e. Store task summary (context for next)    │     │
│  │   f. Unblock dependent tasks                  │     │
│  │   g. Update PRD counters                      │     │
│  │   h. Backup task state to factory-state       │     │
│  │   i. Cleanup worktree (if used)               │     │
│  │   j. Emit notification event                  │     │
│  │                                               │     │
│  │ If failed after max retries:                  │     │
│  │   a. Keep branch (don't delete)               │     │
│  │   b. Task status → "failed"                   │     │
│  │   c. Block dependent tasks                    │     │
│  │   d. Emit notification (task_failed)          │     │
│  │   e. Stop PRD execution chain                 │     │
│  └──────────────────────────────────────────────┘     │
└───────────────────────┬──────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────┐
│            PHASE 3: PRD COMPLETION                    │
│                                                       │
│  After all tasks done:                                │
│  1. PRD alignment review (fresh Reviewer agent)       │
│     Input: original PRD + all task summaries + diffs  │
│     Output: alignment score + gaps found              │
│  2. PRD status → "completed" (or "failed" if gaps)    │
│  3. Generate project summary                          │
│  4. Emit notification (project_completed)             │
│  5. Full backup snapshot                              │
└──────────────────────────────────────────────────────┘
```

### 5.2 Retry Strategy

```
Task retry flow:

  Run N fails
    │
    ├── run_count < max_retries?
    │     │
    │     ├── YES: run_count < escalation_after_retries?
    │     │         │
    │     │         ├── YES: retry with same model + error feedback
    │     │         └── NO:  retry with escalation_model + error feedback
    │     │
    │     └── NO: task status → "failed", stop chain, notify human
    │
    └── Error feedback included in next attempt:
          "Previous attempt failed with: {error}"
          "Review agent feedback: {review.reasons}"
          "Files that were changed: {files_changed}"
```

### 5.3 Dependency Resolution

```
Task dependency resolution:

  Before executing task T:
    1. Parse T.depends_on[]
    2. For each dependency D:
       a. If D is in same project:
          - Check D.status == "done"
       b. If D is cross-project (format: "project-name:task-id"):
          - Lookup project by name
          - Check task D.status == "done"
    3. If all dependencies met → proceed
    4. If any dependency "failed" → mark T as "failed" (cascade)
    5. If any dependency not "done" → T stays "blocked"

  After task T completes:
    1. Find all tasks where depends_on contains T.id
    2. For each dependent task:
       a. Check if ALL its dependencies are now "done"
       b. If yes → change status from "blocked" to "queued"

  Cycle detection (on task creation):
    DFS from new task through depends_on graph
    If visited node seen again → reject with cycle error
```

### 5.4 Timeout Handling

```
Scheduled job: check-timeouts (every 5 minutes)

  1. Find all Runs where status = "running"
  2. For each running Run:
     a. Calculate elapsed = now - run.started_at
     b. Load task = run.task
     c. If elapsed > task.timeout_secs (from project config):
        - Run status → "failed"
        - Run error → "timeout after {elapsed}s"
        - Trigger retry flow (see 5.2)
```

---

## 6. Skills Specification

### 6.1 Existing Skills (Interactive)

These are used by humans in interactive Claude Code sessions. Factory does NOT call these directly.

| Skill | Purpose |
|-------|---------|
| `/grill-me` | Interview about design decisions |
| `/write-prd` | Create YOLO-aware PRD |
| `/prd-to-issues` | Break PRD into GitHub issues |
| `/tdd` | Red-green-refactor implementation |
| `/audit` | Check YOLO conventions |
| `/debrief` | Review agent task execution |
| `/docs` | Discuss YOLO architecture |

### 6.2 Factory Skills (Headless)

New skills designed for headless execution by Factory agents. These output structured data (JSON via `--json-schema`).

#### 6.2.1 factory:plan-tasks

**Purpose:** Break PRD into ordered, dependency-aware tasks.

**Agent:** Planner (Opus)

**Input (via prompt):**
```
PRD title, body, acceptance criteria, design decisions
Project context: CLAUDE.md, entity list, existing code structure
```

**Output (via --json-schema):**
```json
{
  "tasks": [
    {
      "title": "Add RetryPolicy interface to core",
      "spec": "Create RetryPolicy interface in core/retry/...",
      "acceptance_criteria": [
        {"id": "tc-1", "description": "Interface exists"},
        {"id": "tc-2", "description": "Default implementation with exponential backoff"}
      ],
      "branch": "main",
      "sequence": 1,
      "depends_on": [],
      "estimated_complexity": "medium"
    }
  ]
}
```

**Rules:**
- One task = one repo, one branch, one commit
- Tasks must be independently testable
- Specs describe WHAT, not HOW (agent reads code for HOW)
- Dependencies reference task sequence numbers
- Cross-project dependencies use "project-name:sequence" format

#### 6.2.2 factory:implement

**Purpose:** Implement a task following TDD.

**Agent:** Implementer (Sonnet)

**Input (via prompt):**
```
Task spec + acceptance criteria
Implementation plan (from plan step)
Previous task summaries (dependency context)
Framework conventions (CLAUDE.md excerpt)
```

**Behavior:**
1. Read existing code to understand current state
2. Write failing test (red)
3. Write minimal implementation to pass (green)
4. Refactor if needed
5. Repeat for each acceptance criterion
6. Run build + tests to verify

**Output:** Code changes in working directory + commit.

#### 6.2.3 factory:review

**Purpose:** Review implementation against acceptance criteria.

**Agent:** Reviewer (Sonnet, read-only)

**Input (via prompt):**
```
Git diff (what changed)
Task acceptance criteria
YOLO anti-pattern checklist
Previous review feedback (if retry)
```

**Output (via --json-schema):**
```json
{
  "verdict": "pass",
  "criteria_results": [
    {"criteria_id": "tc-1", "passed": true, "comment": "Interface exists at core/retry/policy.go"},
    {"criteria_id": "tc-2", "passed": true, "comment": "Exponential backoff implemented correctly"}
  ],
  "anti_patterns": [],
  "reasons": [],
  "suggestions": ["Consider adding godoc comments to exported types"]
}
```

**Anti-pattern checklist:**
- Hardcoded values that should be configurable
- Missing error handling at system boundaries
- Tests that mock internal code instead of using real implementations
- Code that violates YOLO entity/action patterns
- Missing or incorrect YOLO conventions
- Scope creep — changes beyond what the task spec asked for

#### 6.2.4 factory:audit

**Purpose:** Run `yolo audit` and return structured results.

**Agent:** Auditor (Sonnet, read + bash)

**Input (via prompt):**
```
List of changed files
CLAUDE.md conventions to check against
```

**Behavior:**
1. Run `yolo audit` on the project
2. Parse output for violations
3. Check changed files specifically for convention adherence

**Output (via --json-schema):**
```json
{
  "passed": true,
  "violations": [],
  "warnings": ["TODO in line 45 of core/retry/policy.go"]
}
```

#### 6.2.5 factory:review-prd

**Purpose:** Final PRD alignment check after all tasks complete.

**Agent:** Reviewer (Sonnet, read-only)

**Input (via prompt):**
```
Original PRD (title, body, acceptance criteria)
All task summaries
All git diffs combined
```

**Output (via --json-schema):**
```json
{
  "alignment_score": 0.95,
  "criteria_met": ["ac-1", "ac-2", "ac-3"],
  "criteria_missed": [],
  "scope_drift": [],
  "recommendations": ["Consider adding integration test for retry + dead letter queue interaction"]
}
```

---

## 7. Sentinel System

### 7.1 Purpose

Sentinel is a background agent that watches all registered projects for problems and auto-creates tasks or suggestions.

### 7.2 Watches

| Watch | Trigger | Action | Trust Level |
|-------|---------|--------|-------------|
| **Build Health** | Scheduled (after each task, daily) | Run `go build ./...` | High — auto-create fix task |
| **Test Health** | Scheduled (after each task, daily) | Run `go test ./...` | High — auto-create fix task |
| **Convention Drift** | Scheduled (daily) | Run `yolo audit` | Medium — create suggestion |
| **Dependency Updates** | Scheduled (weekly) | Check `go.mod` for outdated deps | Medium — create suggestion |
| **Security Vulnerabilities** | Scheduled (daily) | Run `govulncheck` | High — auto-create fix task |
| **TODO/FIXME Threshold** | Scheduled (weekly) | Grep for debt markers | Low — create suggestion |
| **Failed Task Analysis** | On task failure (after max retries) | Analyze error patterns | Medium — create suggestion with different approach |
| **Orphaned Runs** | Scheduled (every 5 min) | Find running runs past timeout | High — mark failed, trigger retry |

### 7.3 Trust Levels

```
High trust (auto-create task):
  - Build broken → immediate fix task
  - Tests failing → immediate fix task
  - Security vulnerability → immediate fix task
  - Orphaned runs → mark failed + retry

Medium trust (create suggestion):
  - Convention violations → suggestion
  - Failed task analysis → suggestion with different approach
  - Dependency updates → suggestion

Low trust (create suggestion, low priority):
  - TODO threshold → suggestion
```

### 7.4 Sentinel Schedule

```yaml
sentinel:
  quick_checks:
    trigger: "after_task_completion"  # run after every task
    watches: [build_health, test_health]

  daily_checks:
    schedule: "0 2 * * *"  # 2 AM daily
    watches: [build_health, test_health, convention_drift, security]

  weekly_checks:
    schedule: "0 3 * * 1"  # 3 AM Monday
    watches: [dependency_updates, todo_threshold]

  continuous:
    schedule: "*/5 * * * *"  # every 5 minutes
    watches: [orphaned_runs]
```

---

## 8. Advisor System

### 8.1 Purpose

Advisor is a background agent that analyzes code and execution history to suggest improvements.

### 8.2 Analysis Types

| Analysis | Input | Output |
|----------|-------|--------|
| **Pattern Extraction** | Completed task diffs across projects | "RetryPolicy could be shared in core" |
| **Code Quality** | Codebase scan | "Test coverage dropped from 82% to 71%" |
| **Performance** | Run history (costs, retries, durations) | "Tasks in plugin-X average 3.2 retries — specs may be unclear" |
| **Architecture** | Codebase structure | "12 packages have cyclic dependencies" |
| **Model Optimization** | Run history by model | "Opus tasks in core succeed 95% vs Sonnet 60% — consider Opus for core" |
| **Workflow Optimization** | Step durations, failure points | "Audit step catches 80% of issues — run it before review to save review cost" |

### 8.3 Advisor Schedule

```yaml
advisor:
  post_prd:
    trigger: "after_prd_completion"
    analyses: [pattern_extraction, code_quality]

  weekly:
    schedule: "0 4 * * 1"  # 4 AM Monday
    analyses: [performance, model_optimization, workflow_optimization]

  monthly:
    schedule: "0 4 1 * *"  # 4 AM, 1st of month
    analyses: [architecture]
```

---

## 9. Context & Memory Management

### 9.1 Context Layers

| Layer | What | Managed By | When Loaded |
|-------|------|-----------|-------------|
| **Framework Context** | YOLO conventions, patterns | CLAUDE.md in each repo | Auto-loaded (unless `--bare`) |
| **Project Context** | Architecture, decisions, recent changes | Factory writes to `~/.claude/projects/<repo>/memory/` | Auto-loaded (unless `--bare`) |
| **Task Context** | Previous task summaries, dependency outputs | Factory injects via `--append-system-prompt` | Per task |
| **Step Context** | Plan output, review feedback, error messages | Factory injects via prompt | Per step |

### 9.2 Context Flow Between Tasks

```
Task 1 completes:
  Factory stores task.summary = "Added RetryPolicy interface with exponential backoff"
  Factory stores task.commit_hash = "abc123"
  Factory stores run.files_changed = ["core/retry/policy.go", ...]

Task 2 starts:
  Factory builds context:
    "Previous tasks completed:
     - Task 1 (done): Added RetryPolicy interface (commit abc123)
       Changed: core/retry/policy.go, core/retry/policy_test.go
     
     Current task: Use RetryPolicy in webhook handler
     Acceptance criteria: ..."
```

### 9.3 Session Management

| Session Pattern | When | Why |
|----------------|------|-----|
| **New session** | Review, Audit steps | Fresh context, no bias from implementation |
| **Resume session** | Implement after Plan (same run) | Plan context carries over, saves tokens |
| **Resume with answer** | After question escalation | Agent continues where it left off |
| **Fork session** | Not used in v1 | Future optimization for shared context |

**Naming convention:**
```
factory:{task-id}:{phase}
factory:task-001:plan
factory:task-001:implement
factory:task-001:audit
factory:task-001:review
factory:task-001:question-planner
```

### 9.4 Project Memory Updates

After each task, Factory updates the project's Claude Code memory:

```
~/.claude/projects/<repo>/memory/
├── MEMORY.md                    # index, updated by Factory
├── recent_changes.md            # last 10 task summaries
├── architecture.md              # entity relationships (updated by Advisor)
└── known_issues.md              # from Sentinel findings
```

This benefits both Factory agents AND human interactive sessions — when you open Claude Code in the same repo, you see what Factory has been doing.

---

## 10. Backup System

### 10.1 Architecture

```
Primary: PostgreSQL
Backup:  factory-state git repository (auto-managed)
```

### 10.2 Factory State Repository

```
factory-state/                     # git repo, auto-created on first run
├── projects/
│   ├── yolo-core.yml
│   ├── plugin-webhooks.yml
│   └── app-libromi.yml
├── prds/
│   ├── prd-001-webhook-retry.yml
│   └── prd-002-auth-system.yml
├── tasks/
│   ├── task-001.yml               # includes runs, steps, reviews inline
│   ├── task-002.yml
│   └── task-003.yml
├── questions/
│   └── question-001.yml
├── suggestions/
│   └── suggestion-001.yml
└── snapshots/
    ├── 2026-04-04-daily.yml       # daily full DB dump
    └── 2026-04-05-daily.yml
```

### 10.3 Backup Triggers

| Trigger | What's Backed Up | Frequency |
|---------|-----------------|-----------|
| Task status change | Task YAML (with nested runs/steps/reviews) | On every change |
| PRD status change | PRD YAML | On every change |
| Project config change | Project YAML | On every change |
| Question created/answered | Question YAML | On every change |
| Suggestion created | Suggestion YAML | On every change |
| Daily snapshot | Full DB dump as YAML | Daily at midnight |

### 10.4 Backup Format

Task YAML (nested, one file per task):

```yaml
# tasks/task-001.yml
id: "01JQXYZ..."
prd_id: "01JQABC..."
project_id: "01JQDEF..."
title: "Add RetryPolicy interface to core"
status: "done"
spec: |
  Create RetryPolicy interface in core/retry/ package...
acceptance_criteria:
  - id: "tc-1"
    description: "Interface exists in core/retry/"
  - id: "tc-2"
    description: "Default implementation with exponential backoff"
branch: "main"
model: "sonnet"
sequence: 1
depends_on: []
run_count: 1
max_retries: 3
cost_usd: 0.93
summary: "Added RetryPolicy interface with exponential backoff and jitter"
commit_hash: "abc123def456"
created_at: "2026-04-04T10:00:00Z"
completed_at: "2026-04-04T10:02:30Z"

runs:
  - id: "01JQRUN..."
    agent_type: "implementer"
    model: "sonnet"
    session_id: "claude-uuid-123"
    session_name: "factory:task-001:implement"
    status: "completed"
    cost_usd: 0.84
    tokens_in: 45231
    tokens_out: 12847
    duration_ms: 45000
    num_turns: 12
    commit_hash: "abc123def456"
    files_changed:
      - "core/retry/policy.go"
      - "core/retry/policy_test.go"
    steps:
      - phase: "plan"
        skill: "plan-task"
        status: "completed"
        model: "opus"
        cost_usd: 0.35
      - phase: "implement"
        skill: "tdd"
        status: "completed"
        model: "sonnet"
        cost_usd: 0.30
      - phase: "audit"
        skill: "audit"
        status: "completed"
        model: "sonnet"
        cost_usd: 0.08
      - phase: "review"
        skill: "review-task"
        status: "completed"
        model: "sonnet"
        cost_usd: 0.11
    review:
      verdict: "pass"
      criteria_results:
        - criteria_id: "tc-1"
          passed: true
          comment: "Interface exists at core/retry/policy.go"
        - criteria_id: "tc-2"
          passed: true
          comment: "ExponentialBackoff struct with jitter"
      anti_patterns: []
      suggestions: []
```

### 10.5 Recovery

```bash
# Rebuild DB from backup
factory recover --from /path/to/factory-state

# Recovery reads all YAML files, inserts into DB in dependency order:
# 1. Projects
# 2. PRDs
# 3. Tasks
# 4. Runs + Steps + Reviews
# 5. Questions
# 6. Suggestions
```

---

## 11. MCP Interface

### 11.1 Factory as MCP Server

Factory exposes tools via YOLO's built-in MCP listener. Any MCP client (including interactive Claude Code sessions) can interact with Factory.

### 11.2 MCP Tools

```yaml
tools:
  # Project management
  factory_list_projects:
    description: "List all registered Factory projects with status"
    input: { status: "optional string filter" }
    output: "Project[]"

  factory_get_project:
    description: "Get project details including budget usage"
    input: { project_id: "required" }
    output: "Project with budget summary"

  factory_register_project:
    description: "Register a new project for Factory management"
    input: { name, repo_url, local_path, branch, ... }
    output: "Project"

  # PRD management
  factory_submit_prd:
    description: "Submit a PRD for planning and execution"
    input: { project_id, title, body, acceptance_criteria, design_decisions }
    output: "PRD with generated task list"

  factory_get_prd:
    description: "Get PRD details with task progress"
    input: { prd_id: "required" }
    output: "PRD with tasks and completion status"

  # Task management
  factory_list_tasks:
    description: "List tasks with filters"
    input: { project_id, prd_id, status, branch }
    output: "Task[]"

  factory_get_task:
    description: "Get task details with runs, steps, and reviews"
    input: { task_id: "required" }
    output: "Task with nested runs/steps/reviews"

  # Execution control
  factory_execute:
    description: "Start executing a PRD's tasks"
    input: { prd_id: "required" }
    output: "Execution started confirmation"

  factory_pause:
    description: "Pause execution of a project"
    input: { project_id: "required" }
    output: "Project paused"

  factory_resume:
    description: "Resume paused project execution"
    input: { project_id: "required" }
    output: "Project resumed"

  factory_cancel_task:
    description: "Cancel a queued or running task"
    input: { task_id: "required" }
    output: "Task cancelled"

  # Status
  factory_status:
    description: "Get overall Factory status — running tasks, queue, costs"
    output: "Factory status summary"

  # Questions
  factory_list_questions:
    description: "List open questions from agents"
    input: { status: "optional", project_id: "optional" }
    output: "Question[]"

  factory_answer_question:
    description: "Answer an agent's question"
    input: { question_id, answer }
    output: "Question answered, task resumed"

  # Suggestions
  factory_list_suggestions:
    description: "List pending suggestions from Sentinel/Advisor"
    input: { project_id: "optional", category: "optional", status: "optional" }
    output: "Suggestion[]"

  factory_approve_suggestion:
    description: "Approve a suggestion and convert to task"
    input: { suggestion_id, prd_id: "optional — create new PRD or add to existing" }
    output: "Task created from suggestion"
```

---

## 12. Notification Events

### 12.1 Event Types

Factory emits events via YOLO's event system. `plugin-notifications` handles delivery.

```yaml
events:
  # Task lifecycle
  factory.task.started:
    payload: { task_id, title, project_name }

  factory.task.completed:
    payload: { task_id, title, project_name, cost_usd, duration_ms, commit_hash }

  factory.task.failed:
    payload: { task_id, title, project_name, error, run_count, max_retries }
    severity: warning

  factory.task.needs_review:
    payload: { task_id, title, review_reasons }
    severity: info

  # PRD lifecycle
  factory.prd.planning_complete:
    payload: { prd_id, title, task_count }

  factory.prd.completed:
    payload: { prd_id, title, total_tasks, total_cost_usd, duration }

  factory.prd.failed:
    payload: { prd_id, title, completed_tasks, failed_task }
    severity: error

  # Questions
  factory.question.needs_human:
    payload: { question_id, task_id, body, context }
    severity: warning

  # Budget
  factory.budget.warning:
    payload: { project_name, spent, limit, percentage }
    severity: warning

  factory.budget.exceeded:
    payload: { project_name, spent, limit }
    severity: error

  # Sentinel
  factory.sentinel.build_broken:
    payload: { project_name, error }
    severity: error

  factory.sentinel.security_vuln:
    payload: { project_name, vulnerability, severity }
    severity: critical

  # System
  factory.system.error:
    payload: { error, component }
    severity: error
```

### 12.2 Notification Configuration

Configured via `plugin-notifications` in `app.yml`:

```yaml
notifications:
  channels:
    slack:
      webhook_url: "${SLACK_WEBHOOK_URL}"
      events: ["factory.task.failed", "factory.prd.completed", "factory.question.needs_human", "factory.budget.*", "factory.sentinel.*"]

    email:
      smtp: "${SMTP_URL}"
      to: "dev@example.com"
      events: ["factory.prd.completed", "factory.sentinel.security_vuln"]

    webhook:
      url: "${WEBHOOK_URL}"
      events: ["*"]  # all events
```

---

## 13. CLI Commands

### 13.1 Project Management

```bash
# Register a new project
factory project add \
  --name plugin-webhooks \
  --repo git@github.com:yolo-hq/plugin-webhooks.git \
  --path /repos/plugin-webhooks \
  --branch main \
  --model sonnet

# List projects
factory project list
factory project list --status active

# Get project details
factory project get plugin-webhooks

# Update project config
factory project update plugin-webhooks --auto-merge true --budget-monthly 300

# Pause/resume
factory project pause plugin-webhooks
factory project resume plugin-webhooks

# Archive (stops all execution, marks inactive)
factory project archive plugin-webhooks
```

### 13.2 PRD Management

```bash
# Submit PRD from file
factory prd submit --project plugin-webhooks --file ./prd-webhook-retry.md

# Submit PRD inline
factory prd submit --project plugin-webhooks \
  --title "Webhook retry with exponential backoff" \
  --body "..." \
  --criteria "Retry up to configured max" \
  --criteria "Exponential backoff between retries"

# Generate PRD from one-liner (uses Planner agent)
factory prd generate --project plugin-webhooks \
  "Add retry logic for failed webhook deliveries with exponential backoff"

# List PRDs
factory prd list
factory prd list --project plugin-webhooks --status in_progress

# Get PRD details with task progress
factory prd get prd-001

# Approve PRD (triggers planning if not already done)
factory prd approve prd-001

# Execute PRD (starts task execution)
factory prd execute prd-001
```

### 13.3 Task Management

```bash
# List tasks
factory task list
factory task list --prd prd-001 --status queued
factory task list --project plugin-webhooks

# Get task details
factory task get task-001

# Cancel task
factory task cancel task-001

# Retry failed task manually
factory task retry task-001

# Retry with model override
factory task retry task-001 --model opus
```

### 13.4 Status & Monitoring

```bash
# Overall Factory status
factory status

# Output:
# Factory Status
# ──────────────
# Running:  2 tasks (plugin-webhooks:task-003, yolo-core:task-007)
# Queued:   5 tasks
# Blocked:  3 tasks
# Today:    $4.23 spent
# Month:    $67.89 spent
#
# Active PRDs:
#   prd-001: Webhook retry (4/6 tasks done) — plugin-webhooks
#   prd-002: Auth system (1/8 tasks done) — yolo-core

# Watch mode (live updates)
factory status --watch

# Cost report
factory cost --period month
factory cost --project plugin-webhooks --period week
```

### 13.5 Questions & Suggestions

```bash
# List open questions
factory questions list
factory questions list --status open

# Answer a question
factory questions answer question-001 "Use context.Context for cancellation"

# List suggestions
factory suggestions list
factory suggestions list --category optimization --priority high

# Approve/reject suggestion
factory suggestions approve suggestion-001 --prd prd-003
factory suggestions reject suggestion-001 --reason "Not needed right now"
```

### 13.6 Backup & Recovery

```bash
# Manual backup
factory backup

# Recovery from backup
factory recover --from /path/to/factory-state

# Backup status
factory backup status
```

### 13.7 Sentinel & Advisor

```bash
# Run sentinel checks manually
factory sentinel run --project plugin-webhooks
factory sentinel run --all

# Run advisor analysis manually
factory advisor run --project plugin-webhooks
factory advisor run --analysis performance

# View sentinel/advisor history
factory sentinel history --project plugin-webhooks
factory advisor history --project plugin-webhooks
```

---

## 14. Admin UI

### 14.1 Views

| View | Purpose | Entity |
|------|---------|--------|
| **Dashboard** | Overview: running tasks, costs, recent activity | Aggregate |
| **Projects** | List/manage registered projects | Project |
| **Project Detail** | Config, budget usage, PRDs, tasks | Project |
| **PRDs** | List/filter PRDs by status, project | PRD |
| **PRD Detail** | Task breakdown, progress, cost, timeline | PRD |
| **Tasks** | List/filter tasks with board and table views | Task |
| **Task Detail** | Runs, steps, reviews, questions, git info | Task |
| **Questions** | Open questions needing human answers | Question |
| **Suggestions** | Sentinel/Advisor recommendations | Suggestion |
| **Cost Dashboard** | Cost breakdown by project, PRD, agent type, model | Aggregate |
| **Execution Log** | Timeline of all Factory activity | Run/Step |

### 14.2 Dashboard Layout

```
┌─────────────────────────────────────────────────────────┐
│ YOLO Factory                                             │
├──────────┬──────────────────────────────────────────────┤
│          │                                               │
│ Projects │  Dashboard                                    │
│ PRDs     │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐       │
│ Tasks    │  │Active│ │Queue │ │Done  │ │Failed│       │
│ Questions│  │  2   │ │  5   │ │ 34   │ │  1   │       │
│ Suggest. │  └──────┘ └──────┘ └──────┘ └──────┘       │
│ Costs    │                                               │
│ Logs     │  Running Tasks                                │
│          │  ┌────────────────────────────────────┐      │
│          │  │ plugin-webhooks:task-003            │      │
│          │  │ Step: implement (2m 30s) ████░░ 60% │      │
│          │  │ Model: sonnet | Cost: $0.42         │      │
│          │  └────────────────────────────────────┘      │
│          │                                               │
│          │  Recent Activity                              │
│          │  • task-002 completed ($0.87, 1m 45s)        │
│          │  • task-001 completed ($0.93, 2m 30s)        │
│          │  • prd-001 planning complete (6 tasks)       │
│          │                                               │
│          │  Open Questions (1)                           │
│          │  ┌────────────────────────────────────┐      │
│          │  │ "Should RetryPolicy use context?"   │      │
│          │  │ task-004 | confidence: low          │      │
│          │  │ [Answer] [Dismiss]                  │      │
│          │  └────────────────────────────────────┘      │
│          │                                               │
│          │  Cost This Month: $67.89 / $200.00           │
│          │  ████████████████░░░░░░░░░░░░ 34%            │
│          │                                               │
└──────────┴──────────────────────────────────────────────┘
```

### 14.3 Task Board View

```
┌──────────┬──────────┬──────────┬──────────┬──────────┐
│ Queued   │ Blocked  │ Running  │ Reviewing│ Done     │
├──────────┼──────────┼──────────┼──────────┼──────────┤
│ task-005 │ task-006 │ task-003 │          │ task-001 │
│ task-007 │ task-008 │          │          │ task-002 │
│          │          │          │          │ task-004 │
└──────────┴──────────┴──────────┴──────────┴──────────┘
```

---

## 15. Configuration Reference

### 15.1 app.yml (Full)

```yaml
name: "yolo-factory"
version: "1.0.0"

database:
  primary:
    url: "${DATABASE_URL:postgresql://postgres@localhost:5432/yolo_factory?sslmode=disable}"
    pool:
      max_open: 20
      max_idle: 10
      max_lifetime: "5m"

apps:
  api:
    type: server
    name: YOLO Factory API
    listeners:
      - protocol: http
        addr: ":9000"
        middleware:
          - cors
          - logging
          - recovery
      - protocol: mcp
        transport: sse
        addr: ":3001"
    cors:
      origins: ["http://localhost:3000"]
      methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
      headers: ["*"]
    domains:
      factory:
        entities:
          Project: "*"
          PRD: "*"
          Task: "*"
          Run:
            actions: [create, complete]
            queries: [list, get]
          Step:
            queries: [list, get]
          Review:
            queries: [list, get]
          Question: "*"
          Suggestion: "*"

  worker:
    type: worker
    name: YOLO Factory Worker
    concurrency: 5
    queues:
      - critical
      - default
      - execution
    schedule:
      check-timeouts: "*/5 * * * *"
      sentinel-daily: "0 2 * * *"
      sentinel-weekly: "0 3 * * 1"
      advisor-weekly: "0 4 * * 1"
      advisor-monthly: "0 4 1 * *"
      daily-backup: "0 0 * * *"
    health:
      addr: ":9001"
    shutdown:
      drain: 120s
    domains:
      factory: "*"

  cli:
    type: cli
    name: YOLO Factory CLI
    domains:
      factory: "*"

clients:
  admin:
    port: 3000

# Agent profiles
agent_profiles:
  planner:
    model: opus
    allowed_tools: ["Read", "Glob", "Grep"]
    bare: true
    budget_usd: 1.0
    permission_mode: "auto"
    effort: "high"

  implementer:
    model: sonnet
    allowed_tools: ["Read", "Edit", "Write", "Bash", "Glob", "Grep"]
    bare: true
    budget_usd: 2.0
    permission_mode: "auto"
    effort: "high"

  reviewer:
    model: sonnet
    allowed_tools: ["Read", "Glob", "Grep"]
    bare: true
    budget_usd: 0.5
    permission_mode: "plan"
    effort: "medium"

  auditor:
    model: sonnet
    allowed_tools: ["Read", "Bash", "Glob", "Grep"]
    bare: true
    budget_usd: 0.3
    permission_mode: "auto"
    effort: "medium"

  sentinel:
    model: haiku
    allowed_tools: ["Read", "Bash", "Glob", "Grep"]
    bare: false
    budget_usd: 0.2
    permission_mode: "auto"
    effort: "low"

  advisor:
    model: sonnet
    allowed_tools: ["Read", "Glob", "Grep"]
    bare: false
    budget_usd: 0.5
    permission_mode: "plan"
    effort: "medium"

# Backup
backup:
  state_repo: "${FACTORY_STATE_REPO:}"  # auto-created if empty
  state_path: "${FACTORY_STATE_PATH:./factory-state}"
  auto_backup: true
  daily_snapshot: true

# Sentinel
sentinel:
  enabled: true
  quick_checks_after_task: true

# Advisor
advisor:
  enabled: true

# Notifications (via plugin-notifications)
notifications:
  enabled: true
```

---

## 16. Database Migrations

### 16.1 Migration List

```
001_create_projects.up.sql
002_create_prds.up.sql
003_create_tasks.up.sql
004_create_runs.up.sql
005_create_steps.up.sql
006_create_reviews.up.sql
007_create_questions.up.sql
008_create_suggestions.up.sql
009_create_indexes.up.sql
```

### 16.2 Key Indexes

```sql
-- Task execution order
CREATE INDEX idx_tasks_project_status ON factory_tasks(project_id, status);
CREATE INDEX idx_tasks_prd_sequence ON factory_tasks(prd_id, sequence);
CREATE INDEX idx_tasks_status ON factory_tasks(status);

-- Run lookups
CREATE INDEX idx_runs_task ON factory_runs(task_id);
CREATE INDEX idx_runs_status ON factory_runs(status);

-- Step lookups
CREATE INDEX idx_steps_run ON factory_steps(run_id);

-- Review lookups
CREATE INDEX idx_reviews_run ON factory_reviews(run_id);
CREATE INDEX idx_reviews_task ON factory_reviews(task_id);

-- Question lookups
CREATE INDEX idx_questions_status ON factory_questions(status);
CREATE INDEX idx_questions_task ON factory_questions(task_id);

-- Suggestion lookups
CREATE INDEX idx_suggestions_project_status ON factory_suggestions(project_id, status);

-- Cross-project dependency resolution
CREATE INDEX idx_tasks_id_status ON factory_tasks(id, status);
```

---

## 17. Git Workflow

### 17.1 Per-Task Git Flow

```
WITHOUT WORKTREES (use_worktrees: false):

  1. git checkout {task.branch}         # e.g., "main"
  2. git pull origin {task.branch}
  3. git checkout -b task-{task.id}      # create feature branch
  4. [Agent works here]
  5. git add -A && git commit
  6. git checkout {task.branch}
  7. git merge task-{task.id}
  8. git push origin {task.branch}       # if auto_merge
  9. git branch -d task-{task.id}        # cleanup on success
     # keep branch on failure

WITH WORKTREES (use_worktrees: true):

  1. cd {project.local_path}
  2. git checkout {task.branch}
  3. git pull origin {task.branch}
  4. git worktree add ../project-task-{task.id} -b task-{task.id}
  5. [Agent works in worktree path]
  6. cd {project.local_path}
  7. git merge task-{task.id}
  8. git push origin {task.branch}       # if auto_merge
  9. git worktree remove ../project-task-{task.id}
     # keep worktree on failure for debugging
```

### 17.2 Failed Task Branches

```
On task failure:
  - Branch task-{task.id} is NOT deleted
  - Branch is NOT pushed to remote (unless push_failed_branches: true)
  - Human can:
    a. Inspect: git log task-{task.id}
    b. Fix manually: git checkout task-{task.id}
    c. Retry via Factory: factory task retry {task.id}
    d. Abandon: factory task cancel {task.id} (deletes branch)
```

---

## 18. Error Handling

### 18.1 Error Categories

| Category | Examples | Handling |
|----------|---------|----------|
| **Agent Error** | Claude CLI returns is_error=true | Retry with error feedback |
| **Build Error** | `go build` fails | Retry with build error in prompt |
| **Test Error** | `go test` fails | Retry with test failure in prompt |
| **Audit Error** | `yolo audit` finds violations | Retry with violations in prompt |
| **Review Failure** | Reviewer says "fail" | Retry with review reasons in prompt |
| **Timeout** | Agent exceeds timeout_secs | Kill agent, mark run failed, retry |
| **Budget Exceeded** | Agent exceeds budget cap | Kill agent, mark run failed, notify |
| **Git Error** | Push fails, merge conflict | Mark run failed, notify human |
| **Dependency Cycle** | Circular dependency detected | Reject task creation |
| **Factory Error** | Internal error (DB, worker crash) | Log, notify, mark affected runs as failed |

### 18.2 Retry Feedback Template

When retrying a failed task, the agent receives:

```
## Previous Attempt Failed

### Error
{run.error}

### Review Feedback (if review failed)
{review.reasons}
{review.anti_patterns}

### Files Changed in Failed Attempt
{run.files_changed}

### What To Do Differently
- Address the specific error above
- Do not repeat the same approach that failed
- If the error is unclear, focus on writing tests first to validate your approach
```

---

## 19. Security Considerations

### 19.1 Agent Sandboxing

- Agents run with `--permission-mode auto` — they can edit files and run commands without prompting
- Agents are restricted to the project's `local_path` via `CWD`
- Agents cannot access other project directories
- `--allowedTools` restricts what tools are available per agent type

### 19.2 Secrets

- Factory uses `.env` for secrets (DB URL, API keys, Git tokens)
- Secrets are passed to agent subprocesses via environment variables
- Agents cannot read `.env` directly (it's in Factory's directory, not the project's)
- Never pass secrets in prompts

### 19.3 Git Access

- Factory uses the host's SSH key or Git credential helper
- No tokens stored in Factory DB
- Git operations run as the host user

---

## 20. Observability

### 20.1 Metrics Tracked

| Metric | Granularity | Purpose |
|--------|-------------|---------|
| Cost (USD) | Per step, run, task, PRD, project, monthly | Budget management |
| Tokens (in/out) | Per step, run | Token optimization |
| Duration (ms) | Per step, run, task | Performance tracking |
| Turn count | Per run | Complexity measurement |
| Retry count | Per task | Spec quality indicator |
| Success rate | Per project, agent type, model | Model/skill effectiveness |
| Files changed | Per run | Scope measurement |

### 20.2 Logs

- Factory application logs: structured JSON (YOLO standard)
- Agent session logs: stored by Claude Code at `~/.claude/projects/`
- Execution logs: Factory stores `run.result` and `run.error`
- Git operation logs: captured in step output_summary

---

## 21. Future Enhancements

Features not in current scope but planned for future versions:

| Feature | Description | Priority |
|---------|-------------|----------|
| **Parallel task execution** | Run independent tasks in the same project simultaneously with conflict detection | Medium |
| **Cloud deployment** | Docker image, Kubernetes manifests, cloud workspace provisioning | High |
| **Multi-language support** | Support for non-Go projects (TypeScript, Python, Rust) | Medium |
| **PR-based workflow** | Create PRs instead of direct merge, with review comments | Medium |
| **GitHub issue import** | Auto-create PRDs from labeled GitHub issues | Low |
| **Custom agent types** | Plugin interface for user-defined agent types and workflow steps | Low |
| **Cost prediction** | Estimate cost before execution based on task complexity | Low |
| **A/B model testing** | Run same task with different models, compare results | Low |
| **Team support** | Multiple users, role-based access, shared Factory instance | Medium |
| **Webhooks for external triggers** | Start PRDs from Slack, Linear, Jira, etc. | Medium |
| **Session forking optimization** | Pre-warm context sessions, fork per task for token savings | Medium |
| **Visual diff review** | Admin UI shows git diffs inline with review comments | Low |
| **Auto-rollback** | Detect post-merge failures, auto-revert commits | Medium |
| **TypeScript SDK sidecar** | Optional sidecar for advanced SDK features (multi-turn, custom transport) | Low |

---

## 22. Glossary

| Term | Definition |
|------|-----------|
| **Action** | YOLO pattern for mutations (create, update, execute, cancel) |
| **Agent Profile** | Configuration for an agent type (model, tools, budget) |
| **Acceptance Criteria** | Specific, testable conditions that define "done" |
| **Context Rot** | Quality degradation as an AI agent's context window fills up |
| **Dependency** | A task that must complete before another task can start |
| **Escalation** | Upgrading to a more capable model after repeated failures |
| **Factory State** | Git repository containing YAML backup of all Factory data |
| **Headless Skill** | A skill designed to run without human interaction |
| **MCP** | Model Context Protocol — standard for AI tool integration |
| **PRD** | Product Requirements Document — structured spec for Factory |
| **Run** | One execution attempt of a task |
| **Sentinel** | Background agent that watches for problems |
| **Advisor** | Background agent that suggests improvements |
| **Step** | One workflow phase within a run (plan, implement, test, audit, review) |
| **Worktree** | Git worktree — isolated working directory for a branch |
