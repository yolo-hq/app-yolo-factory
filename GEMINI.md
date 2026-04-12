---
trigger: always_on
---

# GEMINI.md - Yolo Framework

This project uses the **yolo** Go web framework.

## Framework Overview

Yolo is a CQRS-style Go framework with:
- Action-based handlers (Commands/Queries)
- Bun ORM integration
- Built-in validation, auth, and middleware

## Patterns to Follow

### Creating Entities
1. Define struct with `bun.BaseModel`
2. Add `bun` and `json` tags
3. Implement `entity.Entity` interface

### Creating Handlers
1. Create `*Handler` struct with repository
2. Implement HTTP methods (List, Get, Create, Update, Delete)
3. Use `yolohttp.JSON` for responses

### Repository Usage
```go
repo := bun.NewRepository[Entity](db)
items, err := repo.Find(ctx, filter)
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `yolo generate entity Name` | Scaffold entity |
| `yolo generate rules` | Generate AI rules |
| `yolo migrate up` | Run migrations |
