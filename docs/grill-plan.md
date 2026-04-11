# Grill Plan — Factory/Framework Cleanup & Extensions

Analysis of your list: **26 items** grouped into **7 themes**. Each theme is its own grill session — batching unrelated topics causes decision drift.

---

## Session A — TypedData extension to jobs/services/commands
**Items:**
- prd_approve.go fetches Project manually instead of using TypedData
- Extend TypedData pattern to jobs (GraphQL-like, base class, manual resolve `postData.get(entityId)`)
- Services/commands should use same data-loading philosophy as actions

**Key questions:**
- Is `jobs.TypedData[T]` already enough (we shipped it)? Or do you want manual `.get(id)` escape hatch?
- Should services declare `service.TypedData[T]` + caller injects, or auto-load?
- Commands — same pipeline as actions?

---

## Session B — Action conventions cleanup (boilerplate killers)
**Items:**
- `action.NoInput` marker for actions without input (insight_acknowledge.go)
- Unused input fields (insight_dismiss.go `Reason` declared but not used) — lint rule?
- "Hardcoded input values" idea (suggestion_reject.go adding status) — refined via `SetFromInput` + literal Set entries
- Conditional generated fields: `fields.X.When(cond).Value(v)` (suggestion_approve.go)
- question_answer.go not using `SetFromInput` — should it?
- prd_submit.go still uses `actx.Resolve("PRD", res.ID())` — is that legacy?

**Key questions:**
- `NoInput`: marker embed or omit `TypedInput`?
- Conditional fields API shape: `.When(bool).Value(v)` vs `.SetIf(cond, v)` vs helper?
- Lint rule for unused input fields in dev-mode?

---

## Session C — Naming, description, formatting, ordering
**Items:**
- Add `Name()` + `Description()` to all actions/commands/services/jobs
- Formatting command: enforce member order (Name, Description, Execute, ...) run on `yolo dev`
- Remove dead comments project-wide

**Key questions:**
- Auto-fix on save vs warn-only in dev-mode?
- Order convention per type (action vs service vs job)?
- Is `Description()` mandatory or optional? Fail build or warn?

---

## Session D — Commands vs Actions vs MCP
**Items:**
- Factory has commands — do we still need them given MCP?
- Command → can execute action?
- "Action as command" — use action directly as CLI command?

**Key questions:**
- Is command a thin wrapper that just dispatches an action?
- Or drop commands entirely, expose actions via MCP + CLI auto-gen?
- How does this affect `yolo exec` and existing factory commands?

---

## Session E — Read/Write repo codegen (Prisma-like)
**Items:**
- Generate typed read/write repos per entity
- Replace `action.Write[entities.Project](actx).Exec(ctx, write.Update{...})` with something terser

**Key questions:**
- Target API: `repos.Project.Update(actx, id).Status("active").Exec(ctx)`?
- Where does it live — `.yolo/repos/` generated?
- How does this interact with field selection, policies, batch ops?

---

## Session F — Factory migration audit + code quality
**Items:**
- Did we migrate factory to the new framework version? (audit needed)
- Remove `policies/doc.go`
- Move `jsonutil/deps.go` under helpers
- Simplify repeated `json.Unmarshal(payload, &p)` in jobs
- run_complete_test.go — **user said 250 lines, actual 98 lines** — verify claim

**Key questions:**
- Payload unmarshal helper: `jobs.ParsePayload[T](payload)` generic?
- Standard dirs list — is `jsonutil/` approved or should it be `helpers/`?

---

## Session G — Validators + testing philosophy
**Items:**
- Do we still need validators (given TypedData + input struct tags)?
- How to test actions/services/commands/jobs — unified approach?

**Key questions:**
- Collapse validators into action methods or input-level tags?
- Test harness: `yolotest.Action(t).Run(input)` for all types?
- Integration vs unit — which is default?

---

## Recommendation

**Do sessions in this order** (dependency-driven):

1. **Session F** first — audit factory migration state. We can't design extensions on top of unmigrated code.
2. **Session A** — TypedData is the foundation; job/service/command data loading depends on it.
3. **Session B** — action conventions build on A (NoInput, conditional fields).
4. **Session D** — commands-vs-actions decision affects E and G.
5. **Session E** — repo codegen is a big rewrite; needs D settled first.
6. **Session C** — naming/formatting is polish; do after API shape is locked.
7. **Session G** — validators + testing; depends on all above.

**Do NOT batch these into one grill** — 26 decisions in one session = drift. Each session ~20–40 min focused.

## Notes on incorrect claims
- `run_complete_test.go` is **98 lines, not 250** — check if you meant another file.
- `jobs.TypedData[T]` already shipped (PRD #337) — session A is about *extending* it, not creating it.
