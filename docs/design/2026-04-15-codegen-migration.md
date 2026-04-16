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

## Done criteria (per kind)

- Hand dir state matches the final layout in the framework kind file
- All gen orphans under old `.yolo/gen/` removed
- All import sites updated and compiling
- `go build ./...` and `go test ./...` green
- `yolo gen --prune` run on a clean worktree produces no new warnings
- CI passes end-to-end
