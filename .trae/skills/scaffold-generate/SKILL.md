---
name: "scaffold-generate"
description: "Generate a full CRUD feature module (entity/dto/dao/service/controller + route wiring) by mimicking the example user module. Invoke when user says '生成 post 模块' / 'add a Product resource' / '生成文章的 CRUD'."
---

# Scaffold Generate

Generate a new feature module by COPYING THE PATTERN of the existing `user` example module. The `user` module is the single source of truth for conventions — follow it exactly.

## When to invoke

- User asks to generate a resource: "生成 post 模块" / "add Product CRUD" / "做一个订单的增删改查".
- User asks to scaffold an entity.

## Inputs

Ask the user (only if not obvious) for:
1. `resource` — singular, PascalCase, e.g. `Post`, `Product`, `Order`.
2. `table` — DB table name (default: snake_case plural of resource, e.g. `posts`).
3. `fields` — list of `{name, type, json, gorm_tag}`. If user gave a natural-language description, infer types:
   - string → `string` + `gorm:"type:varchar(255)"`
   - int → `int`
   - text/long content → `string` + `gorm:"type:text"`
   - bool → `bool`
   - time → `time.Time`
   - decimal/money → use `decimal.Decimal` (shopspring/decimal)

## Reference module to mimic

Read these files first and mirror their structure, naming, comments style:
- `internal/model/entity/user.go`
- `internal/model/dto/user.go`
- `internal/dao/user.go`
- `internal/service/user.go`
- `internal/controller/user.go`

## Steps

1. **Entity** `internal/model/entity/<snake>.go`:
   - Struct `<Resource>` embedding `BaseModel` (id, created_at, updated_at).
   - Fields from `fields` input, with `json` + `gorm` tags.
   - TableName() returns `table`.
2. **DTO** `internal/model/dto/<snake>.go`:
   - `<Resource>CreateReq`, `<Resource>UpdateReq`, `<Resource>Query`, `<Resource>Resp`.
   - `ToEntity()` / `FromEntity()` conversion helpers (or use mapper in service).
3. **DAO** `internal/dao/<snake>.go`:
   - `<Resource>DAO` struct holding `*gorm.DB`.
   - `Create`, `GetByID`, `List` (with pagination + query filters), `Update`, `Delete`.
   - Constructor `New<Resource>DAO(db *gorm.DB)`.
4. **Service** `internal/service/<snake>.go`:
   - `<Resource>Service` holding the DAO + logger.
   - Business methods delegating to DAO, validating input, mapping entity↔dto.
   - Constructor `New<Resource>Service(dao, logger)`.
5. **Controller** `internal/controller/<snake>.go`:
   - `<Resource>Controller` holding the service.
   - Gin handlers: `Create`, `Get`, `List`, `Update`, `Delete`.
   - Use `pkg/response` for unified responses.
   - Constructor `New<Resource>Controller(svc)`.
   - Expose `RegisterRoutes(rg *gin.RouterGroup)` registering the 5 endpoints under `/<snake_plural>`.
6. **Wire routes** in `internal/router/router.go`:
   - Add import, instantiate controller with its deps, call `RegisterRoutes(apiGroup)`.
7. **Verify**: `go build ./...` and `go vet ./...` must be green.

## Naming conventions

| Layer | File | Struct | Constructor |
|-------|------|--------|-------------|
| entity | `user.go` | `User` | — |
| dto | `user.go` | `UserCreateReq` etc | — |
| dao | `user.go` | `UserDAO` | `NewUserDAO` |
| service | `user.go` | `UserService` | `NewUserService` |
| controller | `user.go` | `UserController` | `NewUserController` |

File name = lowercase singular snake_case of resource. Struct name = PascalCase resource.

## Rules

- Always mimic the `user` example exactly — same package layout, same comment style, same response helpers.
- Never invent new packages; everything goes into the existing layer packages (`controller`, `service`, `dao`, `model/entity`, `model/dto`).
- Always wire the new controller into `router.go` — an unwired controller is a bug.
- Always run `go build ./...` at the end; do not leave the project broken.
- If a field type is ambiguous, ask the user — do not guess silently.
