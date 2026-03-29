# yolo-factory

YOLO app. Domain-driven, action-pipeline architecture.

## Rules
- actx.Resolve("Entity", id) + action.OK() — never return entity data directly
- fields.{key}= for field selection — bare fields= is rejected
- registry.RegisterFilter() for standard list/get. RegisterQueries() for custom reads
- Integration tests only — no mocks, real DB
- Framework first — if YOLO lacks a pattern, build in yolo/ before app code
- action.OK(Extras{}) for non-entity data — action.Success is deprecated

## Structure
server/{domain}/entities/, actions/, inputs/, queries/

## Entities
- Question — server/factory/entities/question.go
- Repo — server/factory/entities/repo.go
- Run — server/factory/entities/run.go
- Task — server/factory/entities/task.go

## Reference
Read docs/*.md in the framework repo for full patterns.

## Verify
go build ./... && go test ./...
