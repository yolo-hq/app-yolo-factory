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
- Project — server/factory/entities/project.go
- PRD — server/factory/entities/prd.go
- Task — server/factory/entities/task.go
- Run — server/factory/entities/run.go
- Step — server/factory/entities/step.go
- Review — server/factory/entities/review.go
- Question — server/factory/entities/question.go
- Suggestion — server/factory/entities/suggestion.go
- Insight — server/factory/entities/insight.go
- LintResult — server/factory/entities/lint_result.go

## Reference
Read docs/*.md in the framework repo for full patterns.

## Verify
go build ./... && go test ./...
