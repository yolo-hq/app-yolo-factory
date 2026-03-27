# FRAMEWORK CONTEXT

Before writing any code, read CLAUDE.md in the repo root. Key rules:

1. All entities MUST embed entity.BaseEntity with TableName() and EntityName() methods
2. All mutations MUST use the action pipeline (action.BaseCreate, action.BaseUpdate, action.BaseDelete, or action.TypedInput with custom Execute())
3. NEVER use raw http.Handler, http.HandlerFunc, or http.ServeMux
4. NEVER use manual SQL (database/sql) — use entity.Repository
5. Follow domain structure: server/{domain}/entities/, actions/, queries/
6. All actions must have Policies() method (yolo.IsAuthenticated(), yolo.Public(), etc.)

# CONTEXT

Start by fetching your own context:

```
cat CLAUDE.md 2>/dev/null
yolo info 2>/dev/null || echo "yolo info not available"
gh issue list --state open --json number,title,body,labels,comments
git log -n 10 --grep="^RALPH:" --format="%H %ad %s" --date=short
cat .ralph/progress.txt 2>/dev/null || echo "No progress yet."
```

Use `yolo info` output to understand project structure (entities, domains, actions) WITHOUT exploring files manually. Only explore files you need to modify.

Use the open issues to find the next task. Use the RALPH commits and .ralph/progress.txt to understand what has already been done.

# TASK SELECTION

Pick the next task. Prioritize tasks in this order:

1. Critical bugfixes
2. Tracer bullets for new features — build a tiny, end-to-end slice of the feature first, then expand it out
3. Polish and quick wins
4. Refactors

Also check `.ralph/questions.md` for resolved questions — if a PARTIAL task now has its questions answered, prioritize completing it.

If all issues are closed, output <promise>COMPLETE</promise>.

# EXPLORATION

Read only the files you need to modify. Do NOT explore the entire repo — use `yolo info` output from the CONTEXT step to understand project structure. Read specific files referenced in the issue description.

# EXECUTION

Complete the task.

If the task involves code changes, follow TDD:
- Write ONE failing test for the first behavior (RED)
- Write minimal code to make it pass (GREEN)
- Refactor if needed
- Repeat for the next behavior
- Do NOT write all tests first then all implementation — one test at a time

If the task is config, docs, or cleanup — skip TDD.

# QUESTIONS & DOUBTS

During exploration or execution, you may encounter doubts. Write a question to `.ralph/questions.md` when you hit:

1. **Ambiguous scope** — the issue could be interpreted multiple ways and the code doesn't clarify
2. **Conflicting patterns** — existing code shows two different approaches for the same thing
3. **Missing dependency** — you need something that doesn't exist and isn't in an open issue
4. **Architectural choice** — multiple valid approaches with different tradeoffs, no clear convention

Do NOT write questions for: implementation details inferable from existing patterns, naming conventions visible in the codebase, anything answered in config, CLAUDE.md, or memory files.

**Always try to self-resolve first.** Read more code, check config, check memory. If you resolve it yourself, log it as `[SELF-RESOLVED]` with your reasoning so the human can audit your assumption.

## Questions file format

If `.ralph/questions.md` doesn't exist, create it. The file uses a counter comment at the top:

```markdown
# Ralph Questions

<!-- next: 2 -->

## Q1 [OPEN] — #<issue-number> <issue title>
**Context:** What you were doing when the doubt arose
**Tried:** What you did to try to resolve it yourself
**Question:** The actual question
**Resolution:** _unresolved_
```

For self-resolved questions, use `[SELF-RESOLVED]` and fill in the Resolution field.

Read the `<!-- next: N -->` counter to get the next question ID, then increment it after writing.

## When blocked by a doubt

Try these in priority order — NEVER stop iterating:

1. **Temp solution** — implement a reasonable default, add `// TODO(ralph): Q<n> — revisit after debrief` in the code, mark the task PARTIAL
2. **Skip the blocked part** — complete everything else in the task that isn't blocked, commit partial work
3. **Move on entirely** — if the whole task is blocked, log the question, leave the issue open, pick a different open issue

## TODO cleanup

If a previously PARTIAL task has its questions resolved (check `.ralph/questions.md` for `[RESOLVED]` entries matching your task), apply the resolution, remove the `// TODO(ralph): Q<n>` comments, and complete the task.

# QUALITY

This is production code. Write it to last:
- Follow existing patterns and conventions in the codebase
- Keep changes small and focused — one logical change per commit
- Do not add speculative features or over-engineer

# FEEDBACK LOOPS

Before committing, run ALL feedback loops:

{{FEEDBACK_LOOPS}}

Do NOT commit if any check fails. Fix issues first.

# COMMIT

Make a git commit. The commit message must:

1. Start with `RALPH:` prefix
2. Include task completed + issue reference (#number)
3. Key decisions made
4. Files changed
5. Blockers or notes for next iteration

Keep it concise.

# PROGRESS

Update .ralph/progress.txt (keep it under 100 lines — trim oldest entries if needed):
- Task completed + issue reference
- Key decisions made
- Files changed
- Blockers or notes for next iteration

Keep entries concise. Sacrifice grammar for concision.

# THE ISSUE

Do NOT close GitHub issues — the orchestrator handles that automatically.

If the task is complete, output `<promise>COMPLETE</promise>` at the end of your response.

If the task is PARTIAL (blocked by a question), do NOT output COMPLETE. Instead:
1. Leave a comment: `RALPH: partial — completed <what>, blocked on <what>. See .ralph/questions.md Q<n>`
2. The issue stays open for the next iteration or after debrief resolves the question.

If the task is not complete for other reasons, leave a comment on the GitHub issue with what was done.

# FINAL RULES

ONLY WORK ON A SINGLE TASK.
