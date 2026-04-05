# YOLO Factory

Autonomous software development engine for the YOLO ecosystem. Submit a PRD, get finished code — planned, implemented, tested, reviewed, and merged.

## Quick Start

```bash
# 1. Start Factory
yolo dev

# 2. Set up the alias
alias yf="yolo run factory"

# 3. Register your projects
yf project:scan --dir ~/projects/yolo-hq

# 4. Submit a PRD
yf prd:submit --project plugin-webhooks --file prd.md

# 5. Execute
yf prd:execute <prd-id>

# 6. Watch
yf status
```

## From Claude Code

```bash
# Add Factory MCP to Claude Code
claude mcp add factory --transport sse --url http://localhost:3001/mcp
```

Then use `/factory:submit` after a `/grill-me` session to send PRDs directly to Factory.

## How It Works

```
PRD → Planner (Opus) breaks into tasks
  → Per task: Plan → Implement (TDD) → Test → Lint → Audit → Review
  → Merge on success → Next task → PRD complete
```

Each task passes 6 quality gates. Program gates (test, lint) cost $0. Agent gates (audit, review) verify acceptance criteria with file:line evidence.

## CLI Reference

```bash
yf project:scan --dir <path>     # auto-discover repos
yf project:list                  # list projects
yf prd:submit --project X --file prd.md  # submit PRD
yf prd:execute <id>              # start execution
yf status                        # current status
yf cost --period month           # cost report
yf task:list --prd <id>          # view tasks
yf prd:diff --id <id>            # combined diff
yf questions:list                # open questions
yf questions:answer <id> "..."   # answer question
yf suggestions:list              # view suggestions
yf insight:list                  # process insights
yf sentinel:run                  # run health checks
```

## Docs

- [Usage Guide](docs/USAGE.md) — end-user documentation
- [Architecture](docs/ARCHITECTURE.md) — how Factory works internally

## Cost

~$0.93 per task. ~$7-8 per PRD (5-8 tasks). ~$100/month for active development.
