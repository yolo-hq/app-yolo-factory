# Codegen Migration — 2026-04-15

Factory-side migration plan for the codegen redesign. The **source of truth** for all design decisions lives in the framework repo:

**→ [`yolo-hq/yolo/docs/design/2026-04-15-codegen-redesign/`](https://github.com/yolo-hq/yolo/tree/0.x/docs/design/2026-04-15-codegen-redesign)**

This document is a thin migration plan: what factory deletes, what it adds to specs, expected touch count, migration order, and blast radius. Do not duplicate framework design here — link instead.

## Problem (factory-specific)

Factory has zombie codegen for 4 of the 14 file kinds. Hand-written files and generated orphans coexist, drifted. Gen files are imported by nobody. The spec pipeline shipped but factory never consumed it.

| Kind | Hand files | Gen orphans | Framework PRD |
|---|---|---|---|
| Entity | `apps/common/factory/entities/` (10 files) | `.yolo/gen/<entity>/entity_gen.go` (10 orphans) | [yolo#563](https://github.com/yolo-hq/yolo/issues/563) |
| Filter | `apps/common/factory/filters/` (10 files) | `.yolo/gen/<entity>/filter_gen.go` (10 orphans) | [yolo#564](https://github.com/yolo-hq/yolo/issues/564) |
| Event | `apps/common/factory/events/` (3 files — `event_types.go`, `emit_generated.go`, `payloads.go`) + `.yolo/events/registry.go` | triple duplication | [yolo#565](https://github.com/yolo-hq/yolo/issues/565) |
| Input | `apps/common/factory/inputs/` (10 files, 0 spec blocks) | none yet | [yolo#566](https://github.com/yolo-hq/yolo/issues/566) |
| Action + Projection | `apps/common/factory/actions/` (16 files ~500 LOC), 10+ projection structs inline | duplicates + hand drift | [yolo#567](https://github.com/yolo-hq/yolo/issues/567) |
| Policy | `apps/common/factory/policies/` (14 files, identical 3-part shape) | 0% gen | [yolo#568](https://github.com/yolo-hq/yolo/issues/568) |
| Service | `apps/common/factory/services/` (16 files, 0 spec blocks, struct-field DI drift) | 0% gen | [yolo#569](https://github.com/yolo-hq/yolo/issues/569) |

## Backup

Tag `v1-pre-codegen-rewrite` pushed 2026-04-15 before any rewrite work. Rollback path exists.

## Migration order

Framework-first per kind. Factory cannot consume a codegen feature that does not exist.

```
yolo#563 (Entity framework)   ──────┐
                                     ├──> factory Entity migration (this repo)
yolo#564 (Filter framework)   ──────┤
                                     ├──> factory Filter migration
yolo#565 (Event framework)    ──────┤
                                     ├──> factory Event migration
yolo#566 (Input framework)    ──────┘
                                     └──> factory Input migration
```

Migration within factory follows the same order — Entity first (others depend on entity field inference and stub file manager shared infrastructure).

## Blast radius estimates

| Kind | Files to delete | Files to add | Spec blocks to add | Expected import-site touches |
|---|---|---|---|---|
| Entity | 10 (hand `entities/`) + 10 (gen orphans under `.yolo/gen/*/entity_gen.go`) | 0 (all gen under `.yolo/factory/entities/`) | 10 `computed:` blocks | ~180 (every action, query, policy importing entities) |
| Filter | 0 from hand (repurposed — old drifted content cleaned) + 10 gen orphans deleted | Tier 2 hand funcs added as needed (~3-5 per entity on average) | 10 `filters:` blocks already present — enriched with `hand` escape rows | ~50 (queries that use filters) |
| Event | 3 hand files + 1 `.yolo/events/registry.go` deleted | 0 hand | 10 `events:` blocks + 1 `specs/events.yml` | ~40 (every action emitting events — migrate to `events.X.Emit(...)` singleton pattern) |
| Input | 10 hand files deleted | 0 hand (unless Tier 2 validators needed) | 10 `inputs:` blocks (NEW — none exist today) | ~60 (every action consuming an input) |

Total expected touch count: **~330 import sites** across actions, queries, policies, services.

## Per-kind factory migration PRDs

Each migration PRD is blocked by its framework counterpart and lives in this repo (`app-yolo-factory`). Created as part of this migration plan.

1. **Factory Entity migration** — [#92](https://github.com/yolo-hq/app-yolo-factory/issues/92), blocked by [yolo#563](https://github.com/yolo-hq/yolo/issues/563)
2. **Factory Filter migration** — [#93](https://github.com/yolo-hq/app-yolo-factory/issues/93), blocked by [yolo#564](https://github.com/yolo-hq/yolo/issues/564)
3. **Factory Event migration** — [#94](https://github.com/yolo-hq/app-yolo-factory/issues/94), blocked by [yolo#565](https://github.com/yolo-hq/yolo/issues/565)
4. **Factory Input migration** — [#95](https://github.com/yolo-hq/app-yolo-factory/issues/95), blocked by [yolo#566](https://github.com/yolo-hq/yolo/issues/566)
5. **Factory Action + Projection migration** — [#96](https://github.com/yolo-hq/app-yolo-factory/issues/96), blocked by [yolo#567](https://github.com/yolo-hq/yolo/issues/567)
6. **Factory Policy migration** — [#97](https://github.com/yolo-hq/app-yolo-factory/issues/97), blocked by [yolo#568](https://github.com/yolo-hq/yolo/issues/568)
7. **Factory Service migration** — #98 (pending file), blocked by [yolo#569](https://github.com/yolo-hq/yolo/issues/569). Covers G11 + G12 retrofits for events moved inline into services/actions.

## Service migration (factory#98 scope)

Source of truth: [`kind-06-service.md`](https://github.com/yolo-hq/yolo/blob/0.x/docs/design/2026-04-15-codegen-redesign/kind-06-service.md).

Delta vs current factory:

- **New spec:** `apps/common/factory/specs/services.yml` with one entry per existing service. ~16 entries, ~800 YAML lines estimated.
- **Hand rewrite:** Each `apps/common/factory/services/<name>.go` becomes `type X struct { gen.X }` + `Execute(ctx, sctx)` + `Validate*` methods. Remove struct-field DI boilerplate (DI now via global `svc` registry gen). Remove `emits:"..."` tag — events move into chain spec.
- **Event inline migration (G12):** Any custom events currently emitted via `service.EmitEvent(ctx, ...)` become `after_success.steps` → `event:` entries in service spec. Action-emitted custom events move into action `events:` block.
- **Chain migration:** Existing service code that manually calls `Actions(actx).X(...)`, `jobs.Defer()`, `Services.X.Execute()` after primary logic — if purely post-Execute orchestration, move to `after_success.steps`. Keep in hand code only when mid-Execute sequencing required.
- **Typed errors:** Each `fmt.Errorf(...)` / `action.Fail(...)` in service migrates to `errors:` list + `s.FailWithCode(gen.X)`. Lint enforces enum coverage.
- **Nested vs top-level rule (S16):** Services called via `sctx.Services.X.Execute()` skip chain automatically. No migration action needed — framework enforces.
- **Projection prefetch (S17):** Identified double-load hot paths from profiling: `run_complete` → `rollback_run` chains reload Run + Tasks. Caller uses `sctx.Prefetch(run, tasks)` to skip re-query.
- **Expected touch count:** ~60 call sites across actions, jobs, other services that invoke services. Plus every event consumer handler referencing custom event names (handler kind still deferred, but registry assembly will catch dangling refs).

Blocked by: yolo#569. Dependencies on earlier migrations: #92 (Entity), #94 (Event — G12 retrofit shared), #95 (Input — validator discovery shared), #96 (Action — chain grammar shared).

Expected blast radius: **~400 LOC Go deleted → ~60 YAML lines + ~200 LOC hand preserved for Execute bodies.**

## Cleanup policy (universal)

- Warn-never-auto-delete. `yolo gen --prune` opt-in to remove orphans
- Pre-check refuses `--prune` if target dir has uncommitted changes
- Git is backup; no `.trash/` dir
- Rename in spec → gen renames, hand warned at next codegen run

## Open risks (factory-specific)

- **Import-site storm** — ~330 touches. Fix with a codemod per kind (grep + replace) before opening migration PR to keep the diff reviewable
- **Test suite breakage** — every integration test uses entities/inputs. Expect widespread test updates
- **Seed data** — current hand fake tags may differ from the spec-derived ones. Verify seed output pre/post
- **CI build time** — `yolo gen` runs at build start. If slow, cache in CI between runs
- **Bootstrap on fresh CI clone** — `.yolo/` empty, codegen must run first. `yolo doctor` should be the entry point in CI as well

## 2026-04-18 retrofit — shared sub-grammars + TypedData migration

Framework locked 4 new global rules + 4 shared sub-grammar docs (G13 Projection, G14 Chain, G15 Bindings, G16 Error codes). Factory migration scope expands to consume them.

### New factory work

| Item | Scope | Blocked by |
|---|---|---|
| TypedData → Projection migration | ~30 actions + ~14 policies — remove `action.TypedData[T]` embed + `*Data` struct; add `projection:` block in action/policy specs; replace `a.Data(actx)` with `actx.Projection.X` | yolo#570 (Projection shared grammar) |
| Action `events:` block → unified chain | ~16 actions — rewrite `events:` block as `after_success:` / `after_fail:` with `event:` / `service:` / `action:` / `job:` items + `with:` | yolo#567 retrofit |
| Action error codes (G16) | Add `errors:` block to actions; align `FailWithCode()` codes | yolo#567 retrofit |
| Policy error codes (G16) | Add `errors:` block to 14 policies; `pctx.Deny(errors.X)` typed | yolo#568 retrofit |
| Service S17 Prefetch cleanup | Drop runtime Prefetch calls where used (audit: likely 0 since not yet implemented) | yolo#569 retrofit |
| Input error codes (I23) | Add `errors:` block to 10 inputs; validator `on_fail:` mapping | yolo#566 retrofit |
| Filter F15 single-file-per-entity | Verify already compliant (likely yes — one filter file per entity convention) | yolo#564 retrofit |

### Codemod: `yolo gen migrate typeddata-to-projection`

Reads existing `type XxxData struct { field tags }` + `action.TypedData[XxxData]` usage, emits YAML projection block per action/policy. Factory runs codemod once; review diff; delete `*Data` structs.

### Touch estimate

| Kind | Additional touches (beyond base migration) |
|---|---|
| Action | ~16 files (chain rewrite + projection block + errors block) |
| Policy | ~14 files (projection block + errors block) |
| Service | audit for Prefetch (expect 0) |
| Input | ~10 files (errors block) |
| Filter | ~0 (F15 already implicit) |

Total additional: ~40 file touches on top of base migration. Spec-heavy, Go-light.

### Migration PRDs

- factory#99 — TypedData → Projection grammar migration (blocked by yolo#570)
- [factory#100](https://github.com/yolo-hq/app-yolo-factory/issues/100) — Handler migration (Consumer → Handler rename + convert consumers); blocked by yolo#571
- Existing PRDs #96-#98 updated with retrofit scope comments.

### 2026-04-18b — Handler kind migration (factory#100)

**Scope:** Migrate factory's `event.Consumer` implementations to new `event.Handler` pattern. Framework kind grilled (kind-08-handler.md, yolo#571).

**Audit targets:**
- Find all `event.Consumer` impls in factory (`services/*_consumer.go`, `events/consumers/*`, or scattered)
- Likely 5-10 files based on typical domain app (H74)

**Code migration per consumer:**
- Remove `event.Consumer` interface impl
- Embed `gen.{HandlerName}` struct
- Update `Execute` signature to typed `Execute(hctx handler.Context, event *gen.{EventName}) error`
- Replace inline projection (ConsumerFields-style) with `hctx.Projection.X`
- Add spec entry in `specs/{domain}/handlers.yml`
- Move file to `apps/common/{domain}/handlers/`

**Codemod:** `yolo gen migrate consumer-to-handler` — mechanical rewrite + scaffold handlers.yml entry.

**Framework entities created automatically on first `yolo migrate up` post-upgrade:**
- `handler_executions` (dedup + log)
- `handler_dead_letter` (failed jobs)
- `event_log` (optional, opt-in via `handler.event_log_enabled: true` in app.yml)

**Factory app.yml additions:**

```yaml
handler:
  dead_letter_retention: 30d          # default
  execution_log_retention: 7d         # default
  auto_cleanup: true
  event_log_enabled: false            # optional
```

**Additional touches:**

| Item | Estimate |
|---|---|
| Handler spec entries | ~5-10 YAML |
| Go code migration | ~200-400 LOC |
| Test updates | ~100-300 LOC |
| `Consumer` → `Handler` call-site rename | ~20-40 |
| Total | ~500-800 LOC |

**Merge strategy:** Coordinate with factory#94 (Event kind migration) — both touch event pipeline. Event PR first (contract side), Handler PR second (consumer side). Or bundle if practical.

### Blast radius: no new import storms

All retrofits are either spec-level (YAML) or API-substitution inside existing Go files. No new import sites beyond what base migration already touches.

## Done criteria (per kind)

- Hand dir state matches the final layout in the framework kind file
- All gen orphans under old `.yolo/gen/` removed
- All import sites updated and compiling
- `go build ./...` and `go test ./...` green
- `yolo gen --prune` run on a clean worktree produces no new warnings
- CI passes end-to-end
