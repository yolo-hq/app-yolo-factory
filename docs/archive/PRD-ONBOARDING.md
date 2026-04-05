# PRD: Factory Onboarding, MCP Integration & Documentation

## Problem Statement

Factory v2 is built (10 entities, 12 services, 31 CLI commands, quality gates, self-improvement). But a new user can't start using it without reading internal design docs. There's no README, no guided onboarding, no way to connect Claude Code to Factory, and the `/write-prd` → Factory flow is disconnected (requires manual copy-paste via CLI).

The docs are scattered: SRS.md (2000 lines), DESIGN.md, two stale PRD files, and a USAGE.md that mixes internal details with user-facing instructions.

## Solution

1. **`project:scan` command** — auto-discovers git repos in a directory, registers them as Factory projects
2. **`prd:diff` command** — shows combined diff for a completed PRD (what Factory actually changed)
3. **Smart MCP tools** — `factory_submit_prd` and `factory_scan_projects` for Claude Code integration
4. **`/factory:submit` skill** — interactive flow: extract PRD from conversation → submit to Factory → show task breakdown → execute
5. **README.md** — intro, quick start, `yf` alias
6. **Polished USAGE.md** — end-user focused, no internals
7. **ARCHITECTURE.md** — merged key sections from SRS + DESIGN
8. **Archive stale docs**

## User Stories

1. As a developer, I want to run `yf project:scan --dir ~/projects/yolo-hq` to register all my repos at once, so that I don't type 15 `project:add` commands.
2. As a developer, I want to use `/factory:submit` in Claude Code after a `/grill-me` session, so that the PRD goes directly to Factory without copy-paste.
3. As a developer, I want Factory's MCP tools available in Claude Code, so that I can ask "What's Factory doing?" and get real-time answers.
4. As a developer, I want `yf prd:diff <id>` to see everything Factory changed for a PRD, so that I can review the combined output.
5. As a developer, I want a README that shows me how to start in under 2 minutes.
6. As a developer, I want the USAGE.md to tell me what to DO, not how Factory works internally.
7. As a developer, I want the `yf` alias documented so I don't type `yolo run factory` every time.
8. As a developer, I want `/factory:submit` to show me the task breakdown before executing, so that I can review and approve.
9. As a developer, I want `/factory:submit` to work without a GitHub issue — PRD goes directly to Factory's DB.
10. As a developer, I want MCP to auto-detect which project a PRD belongs to based on the current working directory.

## YOLO Design

### Entities
No new entities. Uses existing: Project, PRD, Task.

### Actions
No new actions. `project:scan` and `prd:diff` are CLI commands, not HTTP actions.

### New Commands

**`project:scan`** — `server/factory/commands/project_scan.go`
```
Name: "project:scan"
Input:
  Dir    string  flag:"dir" validate:"required" usage:"Directory to scan for git repos"
  DryRun bool    flag:"dry-run" usage:"Show repos without registering"
  Branch string  flag:"branch" usage:"Default branch for all (default: main)"
  Model  string  flag:"model" usage:"Default model for all (default: sonnet)"

Behavior:
  1. Walk Dir, find directories containing .git/ AND go.mod
  2. For each: extract repo name from go.mod module path or directory name
  3. Extract remote URL from `git remote get-url origin`
  4. Print table: name, path, remote, branch
  5. If not --dry-run: ask confirmation, create Project entities
  6. Skip repos already registered (match by name or local_path)
```

**`prd:diff`** — `server/factory/commands/prd_diff.go`
```
Name: "prd:diff"
Input:
  PRDID string  (positional, required)

Behavior:
  1. Load PRD + all tasks with status "done"
  2. For each task: load commit_hash
  3. In the project repo, run `git diff <first-commit>^..<last-commit>`
  4. Print combined diff
  5. Also print summary: files changed, total additions/deletions
```

### MCP Tools (custom, not auto-generated CRUD)

**`factory_submit_prd`** — registered as custom MCP tool
```
Input:
  project_name    string  (required OR auto-detect from cwd)
  title           string  (required)
  body            string  (required)
  acceptance_criteria  []string  (required)
  design_decisions     []string  (optional)
  auto_execute    bool    (default: false — plan only, don't execute yet)

Behavior:
  1. Resolve project by name (or by matching cwd to project.local_path)
  2. Validate: project active, criteria non-empty
  3. Create PRD entity (status: draft)
  4. If auto_execute: dispatch PlanPRDJob
  5. Return: PRD ID, status, project name
```

**`factory_scan_projects`** — registered as custom MCP tool
```
Input:
  dir    string  (required)
  register  bool  (default: false — dry run by default)

Behavior:
  Same as project:scan command but via MCP
  Returns: list of found repos with registration status
```

These custom MCP tools need to be registered in YOLO's MCP listener. Check how YOLO registers custom MCP tools (beyond auto-generated entity CRUD).

If YOLO doesn't support custom MCP tools yet: register them as actions on a virtual "Factory" entity, or as custom HTTP handlers exposed via MCP.

### Skill: `/factory:submit`

**Location:** This skill should live in the Factory repo as a Claude Code plugin skill, not in DevKit.

**File:** `skills/factory-submit/SKILL.md`

```markdown
---
name: factory:submit
description: Submit a PRD to YOLO Factory for autonomous execution
---

Extract a PRD from the current conversation and submit it to Factory.

## Steps

1. Ask which project this is for (or detect from cwd)
2. Extract from conversation:
   - Title (short, descriptive)
   - Body (what to build and why)
   - Acceptance criteria (testable conditions)
   - Design decisions (if discussed)
3. Show the structured PRD to the user for review
4. Submit to Factory via MCP: factory_submit_prd
5. Trigger planning (Planner breaks into tasks)
6. Show task breakdown to user
7. Ask: "Execute? [Y/n]"
8. If yes: trigger execution via MCP

## Important
- Do NOT create a GitHub issue — Factory PRD entity is the source of truth
- Acceptance criteria must be specific and testable
- If /grill-me or /write-prd was used earlier in conversation, extract from that context
- If no MCP connection, fall back to showing the CLI command to run manually
```

### app.yml Changes
If custom MCP tools need registration:
```yaml
apps:
  api:
    listeners:
      - protocol: mcp
        transport: sse
        addr: ":3001"
        tools:
          - factory_submit_prd
          - factory_scan_projects
```

(Depends on how YOLO registers custom MCP tools — may need framework exploration)

## Implementation Decisions

### Modules

**1. project:scan command** — `server/factory/commands/project_scan.go`
- Walk filesystem, find .git + go.mod
- Parse go.mod for module name
- Parse `git remote get-url origin` for URL
- Dedup against existing projects
- Deep module: lots of edge cases (missing remote, bare repos, nested repos) behind simple `scan --dir` interface

**Tests:** Create temp dir with mock git repos, verify discovery.

**2. prd:diff command** — `server/factory/commands/prd_diff.go`
- Load completed tasks, collect commit hashes
- Run `git diff` between first and last commit
- Format output

**Tests:** Use temp git repo with commits, verify diff output.

**3. MCP tools** — depends on YOLO's custom MCP tool registration
- If YOLO supports custom MCP handlers: register as MCP tools
- If not: register as actions on entities, exposed via standard MCP

**4. /factory:submit skill** — `skills/factory-submit/SKILL.md`
- Pure markdown, no code
- Claude Code loads it as context when user types `/factory:submit`
- References MCP tools for submission

**5. Documentation**
- README.md — repo root, 50 lines max
- USAGE.md — polished, end-user focus, ~300 lines
- ARCHITECTURE.md — merged from SRS+DESIGN, ~500 lines
- Archive: move SRS.md, DESIGN.md, PRD.md, PRD-QUALITY.md to docs/archive/

### MCP Config for Claude Code

User adds to their `.claude/settings.json` or project `.claude/settings.local.json`:
```json
{
  "mcpServers": {
    "factory": {
      "type": "sse",
      "url": "http://localhost:3001/mcp"
    }
  }
}
```

Or via CLI:
```bash
claude mcp add factory --transport sse --url http://localhost:3001/mcp
```

## Testing Decisions

- `project:scan` — integration test with temp filesystem + git repos
- `prd:diff` — integration test with temp git repo + commits
- MCP tools — test via HTTP calls to MCP endpoint
- Skill — manual test (markdown, no code to test)
- Docs — review for accuracy

## Out of Scope

- `factory init` command (yolo dev handles setup)
- `/factory:status` skill (MCP tools handle it naturally)
- Modifying `/write-prd` DevKit skill
- GitHub issue integration
- Web UI for onboarding (admin UI already exists)

## Further Notes

- `yf` alias: `alias yf="yolo run factory"` — documented in README
- The `/factory:submit` skill depends on Factory MCP being reachable. If MCP is down, skill falls back to showing the CLI command.
- PRDs submitted via MCP skip GitHub issues entirely. Factory PRD entity is the source of truth.
- The skill extracts PRD from conversation context — works best after `/grill-me` or detailed design discussion.
- `project:scan` is idempotent — running twice skips already-registered repos.
