---
description: Yolo Go framework patterns and conventions
globs: ["**/*.go"]
---

# Yolo Framework Rules

This project uses the **yolo** Go framework for building web APIs.

## Core Patterns

### CQRS Actions
- **Commands**: Mutate state (Create, Update, Delete)
- **Queries**: Read state (Get, List, Find)

```go
// Command handler example
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        yolohttp.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    // ... handle command
}
```

### Entity Definition
```go
type Entity struct {
    bun.BaseModel `bun:"table:entities"`
    ID        string    `bun:",pk,type:uuid,default:gen_random_uuid()"`
    CreatedAt time.Time `bun:",nullzero,default:current_timestamp"`
}
```

### Repository Pattern
- Use `bun.Repository[T]` for data access
- Inject via constructor, not globals

### Error Handling
- Use `yolohttp.Error(w, msg, code)` for HTTP errors
- Use `yolohttp.JSON(w, code, data)` for JSON responses

## Project Structure
```
internal/
  entity/      # Domain entities
  repository/  # Data access
  handler/     # HTTP handlers
cmd/
  api/         # Main application
```
