# AGENTS.md

This project uses the **yolo** Go framework.

## Quick Reference

| Task | Pattern |
|------|---------|
| New entity | `yolo generate entity Name` |
| HTTP handler | Implement on `*Handler` struct |
| Validation | Use `validate` struct tags |
| Database | Bun ORM with repositories |

## Key Imports
```go
import (
    "github.com/yolo-hq/yolo/core/bun"
    yolohttp "github.com/yolo-hq/yolo/core/http"
    "github.com/yolo-hq/yolo/core/web/middleware"
)
```

## Commands
- `yolo generate entity User` - Generate entity scaffold
- `yolo migrate up` - Run migrations
- `yolo openapi` - Generate OpenAPI spec
