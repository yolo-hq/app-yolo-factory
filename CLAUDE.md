# yolo-factory

YOLO app. Domain-driven, action-pipeline architecture.

## Rules
- actx.Resolve("Entity", id) + action.OK() — never return entity data directly
- fields.{key}= for field selection — bare fields= is rejected
- registry.RegisterFilter() for standard list/get
- Integration tests only — no mocks, real DB
- Framework first — if YOLO lacks a pattern, build in yolo/ before app code
- action.OK(Extras{}) for non-entity data
- Use status constants from entities/status.go — never string literals
- Use shared helpers from services/helpers.go — never duplicate functions

## Structure
server/factory/entities/, actions/, inputs/, services/, jobs/, commands/, skills/, lint/, events/

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

## Status Constants
All status/enum values in entities/status.go. Import and use constants.

## Quality Gates
Orchestrator runs: plan → implement → test (go build+test+vet) → lint (factory lint) → audit → review

## Verify
go build ./... && go test ./... && go vet ./...
