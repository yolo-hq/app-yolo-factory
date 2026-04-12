# Yolo Framework Rules

## Overview
Go web API framework with CQRS patterns.

## Quick Patterns

### New Entity
```bash
yolo generate entity EntityName
```

### Handler Structure
```go
type Handler struct {
    repo *repository.EntityRepository
}

func NewHandler(repo *repository.EntityRepository) *Handler {
    return &Handler{repo: repo}
}
```

### Response Helpers
- `yolohttp.JSON(w, 200, data)`
- `yolohttp.Error(w, "msg", 400)`
