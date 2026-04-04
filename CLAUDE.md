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
- Step — server/factory/entities/step.go
- Insight — server/factory/entities/insight.go
- Project — server/factory/entities/project.go
- Suggestion — server/factory/entities/suggestion.go
- Task — server/factory/entities/task.go
- LintResult — server/factory/entities/lintresult.go
- PRD — server/factory/entities/prd.go
- Question — server/factory/entities/question.go
- Review — server/factory/entities/review.go
- Run — server/factory/entities/run.go

## Reference
Read docs/*.md in the framework repo for full patterns.

## Verify
go build ./... && go test ./...
