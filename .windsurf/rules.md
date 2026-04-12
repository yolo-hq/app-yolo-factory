# Windsurf Rules - Yolo Framework

## Project Type
Go web API using yolo framework

## Patterns

### Entity
```go
type Entity struct {
    bun.BaseModel
    ID string `bun:",pk,type:uuid"`
}
```

### Handler
```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    // Decode, validate, save, respond
}
```

### Key Packages
- `github.com/yolo-hq/yolo/core/bun` - Repository
- `github.com/yolo-hq/yolo/core/http` - HTTP helpers
- `github.com/yolo-hq/yolo/core/web/middleware` - Middleware
