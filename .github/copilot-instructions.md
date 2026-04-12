# GitHub Copilot Instructions

This project uses the **yolo** Go framework for web APIs.

## Coding Patterns

### Entity Creation
- Embed `bun.BaseModel` in structs
- Use UUID primary keys with `gen_random_uuid()`
- Add `validate` tags for validation

### Handler Pattern
- Inject repositories via constructor
- Use `yolohttp.Error()` and `yolohttp.JSON()`
- Validate input before processing

### Imports
```go
import (
    "github.com/yolo-hq/yolo/core/bun"
    yolohttp "github.com/yolo-hq/yolo/core/http"
)
```

## Project Layout
- `internal/entity/` - Domain models
- `internal/handler/` - HTTP handlers
- `internal/repository/` - Data access
