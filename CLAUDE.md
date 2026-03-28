# yolo-factory

YOLO framework app — AI agent task orchestration platform.

- Entities embed entity.BaseEntity, actions use action.BaseCreate/TypedInput
- Bootstrap: yolo.MustRunBinary() in main.go
- Structure: server/factory/entities/, actions/, inputs/, filters/, jobs/
- Run `go build ./...` and `go test ./...` before committing
- Framework first: if a pattern doesn't exist in YOLO, build it in the framework
