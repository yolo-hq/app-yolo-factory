# YOLO Factory — Architecture

Internal architecture reference. For usage, see [USAGE.md](USAGE.md). Full details in [archive/SRS.md](archive/SRS.md) and [archive/DESIGN.md](archive/DESIGN.md).

## System Architecture

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
│  │  Go CLI Wrapper (pkg/claude/)         │   │
│  │    └── spawns claude CLI subprocess   │   │
│  │    └── parses JSON output             │   │
│  └──────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
        │
        ▼
┌─────────────────┐
│  Claude Code CLI │ (--bare --output-format json)
└─────────────────┘
        │
        ▼
┌─────────────────┐
│  Target Repos    │
└─────────────────┘
```

## Entity Relationships

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

Suggestion (0..1) ── (1) Task  (if converted)
```

**10 entities:** Project, PRD, Task, Run, Step, Review, Question, Suggestion, Insight, LintResult.

## Workflow

```
1. SUBMIT
   Human writes PRD with acceptance criteria
   PRD status: draft → approved

2. PLAN
   Planner agent (Opus) reads PRD + codebase
   Creates ordered tasks with dependencies
   PRD status: approved → planning → in_progress

3. PER TASK (sequential within project):
   a. Plan      — Opus reads code, creates implementation plan
   b. Implement — Sonnet writes code via TDD (resumes plan session)
   c. Test      — go build + go test (shell, $0)
   d. Lint      — AST + grep checks (program, $0)
   e. Audit     — Sonnet checks YOLO conventions
   f. Review    — Sonnet verifies acceptance criteria with file:line evidence

   Retry on failure → escalate model after N retries → fail after max

4. COMPLETE
   Merge to target branch → push → unblock dependent tasks
   After all tasks → PRD alignment review
   PRD status: in_progress → completed
```

## Agent Types

| Agent | Purpose | Model | Budget |
|-------|---------|-------|--------|
| Planner | Break PRD into tasks, create plans | Opus | $1.00 |
| Implementer | Write code via TDD | Sonnet | $2.00 |
| Reviewer | Verify acceptance criteria | Sonnet | $0.50 |
| Auditor | Check YOLO conventions | Sonnet | $0.30 |
| Sentinel | Auto-detect breaks, create fix tasks | Haiku | $0.20 |
| Advisor | Suggest optimizations, refactoring | Sonnet | $0.50 |

## Quality Gates Pipeline

```
Plan → Implement → Test → Lint → Audit → Review
                   ^^^^   ^^^^
                   $0     $0     (program gates)
```

| Gate | Type | Cost | Catches |
|------|------|------|---------|
| Test | Program | $0 | Build failures, test failures |
| Lint | Program | $0 | Swallowed errors, shell injection, stubs, duplicates, status literals |
| Audit | Agent | ~$0.08 | YOLO convention violations |
| Review | Agent | ~$0.15 | Acceptance criteria not met, anti-patterns, scope creep |

Program gates catch ~80% of issues at zero token cost.

## PRD State Machine

```
         submit
(none) ────────► draft
                   │ approve
                   ▼
                approved
                   │ plan-tasks
                   ▼
                planning
                   │ tasks created
                   ▼
               in_progress
                 │     │
                 │     │ all tasks done
                 │     ▼
                 │  completed
                 │
                 │ task failed after max retries
                 ▼
                failed
```

## Task State Machine

```
              create
(none) ────────────► queued (no deps) OR blocked (has unmet deps)

blocked ── all deps done ──► queued
queued ─── picked ─────────► running
running ── steps pass ─────► reviewing
reviewing ─ review passes ─► done
          └ review fails ──► running (retry)
running ── step fails ─────► running (retry) OR failed (max retries)

Any state ── human cancels ► cancelled
```

## Design Decisions

| # | Decision | Reasoning |
|---|----------|-----------|
| D1 | Sequential per project | Parallel tasks cause git/merge conflicts |
| D2 | No GitHub issues | Factory tasks are source of truth; structured data GitHub can't enforce |
| D3 | Go CLI wrapper | No Go SDK; CLI --output-format json provides everything needed |
| D4 | YOLO app | Dogfooding; gets entities, actions, admin, MCP, CLI for free |
| D5 | Worktrees optional, default off | go.work replace directives break with worktrees |
| D6 | YAML backup in git repo | Human-readable, diffable, portable, full audit trail |
| D7 | Implementation != Review | Fresh context produces honest reviews |
| D8 | Structured output via --json-schema | Eliminates fragile text parsing |
| D9 | Model per step, not per task | Opus for planning, Sonnet for implementation optimizes cost |
| D10 | Session resume Plan→Implement | Carries plan context without re-injection |
| D11 | Question escalation chain | Auto-resolve → Planner → Human |
| D12 | Sentinel auto-creates, Advisor suggests | Different trust levels for different signal quality |
| D13 | Task summaries as context | Cheaper than re-analyzing git diffs |
| D14 | No sub-tasks | Steps handle phases; sub-tasks create competing hierarchies |
| D15 | Auto-merge by default | Trust the automated gates |

## Full Reference

- [archive/SRS.md](archive/SRS.md) — complete requirements specification
- [archive/DESIGN.md](archive/DESIGN.md) — complete design document with prompt templates
- [archive/PRD-ENGINE.md](archive/PRD-ENGINE.md) — original Factory v2 PRD
- [archive/PRD-QUALITY.md](archive/PRD-QUALITY.md) — quality system PRD
