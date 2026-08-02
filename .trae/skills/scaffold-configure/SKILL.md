---
name: "scaffold-configure"
description: "Enable or disable optional modules (auth/cache/queue/websocket/upload) in the Go scaffold by wiring or unwiring their routes and deps. Invoke when user says '启用 auth 模块' / 'add cache module' / '不需要队列' / 'configure modules'."
---

# Scaffold Configure

The scaffold ships with a set of optional modules. This skill turns them on or off cleanly, so the project only carries what it actually needs.

## Module registry

| Module | Location | Deps | Route prefix |
|--------|----------|------|--------------|
| auth | `internal/modules/auth` | `golang-jwt/v5` | `/auth` |
| cache | `internal/modules/cache` | `go-redis/v9` | (service only) |
| queue | `internal/modules/queue` | `hibiken/asynq` | (service only) |
| websocket | `internal/modules/websocket` | `gorilla/websocket` | `/ws` |
| upload | `internal/modules/upload` | (local fs) | `/upload` |

Core (always on): `config`, `router`, `server`, `middleware`, `pkg/*`, example `user` module.

## When to invoke

- User lists desired modules: "我需要 auth 和 cache" / "I want auth + upload, no queue".
- User asks to remove a module: "去掉 websocket" / "remove queue".
- User asks what modules are available: "有哪些可选模块".

## Steps to ENABLE a module

1. Confirm the module exists in `internal/modules/<name>/`. If it doesn't, tell the user and stop.
2. Add its dependency to `go.mod` (`go get <dep>`), then `go mod tidy`.
3. Wire it into `internal/router/router.go`:
   - Add import.
   - Call its `RegisterRoutes(r *gin.RouterGroup)` (or equivalent) inside the appropriate group.
4. If it needs config keys, add them to `configs/config.yaml` and the config struct in `internal/config/config.go`.
5. `go build ./...` and `go vet ./...` — must be green.

## Steps to DISABLE a module

1. Remove its wiring from `internal/router/router.go` (import + registration call).
2. Optionally delete `internal/modules/<name>/` directory if user wants it gone physically.
3. Remove its config keys if no longer referenced.
4. `go build ./...` — must still compile.

## Rules

- Always verify build after wiring changes.
- Never leave dangling imports.
- If a module is missing, list available modules and ask which to create.
- Keep router.go changes minimal and grouped by module.
