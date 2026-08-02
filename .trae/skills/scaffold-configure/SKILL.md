---
name: "scaffold-configure"
description: "Enable or disable optional modules (auth/cache/queue/websocket/upload) in the Go scaffold by wiring or unwiring their routes and deps. Invoke when user says '启用 auth 模块' / 'add cache module' / '不需要队列' / 'configure modules'."
---

# Scaffold Configure

The scaffold ships with a set of optional infrastructure modules. This skill turns them on or off cleanly, so the project only carries what it actually needs.

## Architecture overview

The scaffold has NO `internal/modules/` directory. Optional infrastructure lives under `pkg/service/*` as reusable service kits, and business modules live under `mod/<name>/`. Wiring happens in `main.go` via `restkit.AddActions(mod.<name>.All()...)`.

### Optional service kits (in `pkg/service/`)

| Kit | Location | Config keys (under `pkg/cli/configkey/`) | Purpose |
|-----|----------|------------------------------------------|---------|
| rediskit | `pkg/service/rediskit/` | `redis.go` | Redis client + cache helpers |
| cachekit | `pkg/service/cachekit/` | `cache.go` | Cache abstraction (wraps rediskit) |
| jwtkit | `pkg/service/jwtkit/` | `jwt.go` | JWT token sign/verify |
| mqttkit | `pkg/service/mqttkit/` | `mqtt.go` | MQTT client + pub/sub |
| cronkit | `pkg/service/cronkit/` | — | Cron scheduler |
| excelkit | `pkg/service/excelkit/` | — | Excel read/write |
| pdfkit | `pkg/service/pdfkit/` | — | PDF generation |
| aikit | `pkg/service/aikit/` | — | LLM chat model client |
| storagekit | `pkg/service/storagekit/` | `minio.go` | Local/MinIO storage |

### Core (always on)

- `pkg/cli` — cobra CLI framework + configkit + tag
- `pkg/class` — base types (class.String/Int64/Time/Decimal/etc.) + exception + httpconst
- `pkg/library` — pure utility libs (stringkit, jsonkit, cryptokit, filekit, etc.)
- `pkg/service/restkit` — HTTP server + router + middleware + openapi + context
- `pkg/service/sqlkit` — generic DAO + datasource + transaction
- `pkg/service/configkit` — config loader (file + env)
- `pkg/service/logkit` — logger
- `mod/user` — example business module (user/role/department/smscode)

## When to invoke

- User lists desired modules: "我需要 redis 和 jwt" / "I want auth + cache, no mqtt".
- User asks to remove a kit: "去掉 mqtt" / "remove queue".
- User asks what modules are available: "有哪些可选模块".

## Steps to ENABLE a kit

1. **Verify the kit exists** in `pkg/service/<kit>/`. If not, tell the user it's unavailable and list what exists.
2. **Add dependency** to `go.mod` if the kit needs an external lib:
   - rediskit → `go get github.com/redis/go-redis/v9`
   - jwtkit → `go get github.com/golang-jwt/jwt/v5`
   - mqttkit → `go get github.com/eclipse/paho.mqtt.golang`
   - sqlkit (postgres) → `go get github.com/jackc/pgx/v5`
   Then `go mod tidy`.
3. **Add config keys** if needed:
   - Add constants to `pkg/cli/configkey/<kit>.go` (e.g. `RedisHost`, `RedisPort`).
   - Add default values to the config file the project uses (configkit loads from file/env; check `pkg/cli/configkey/` for available keys).
4. **Initialize the kit** at startup if it needs init (e.g. rediskit reads config keys automatically; jwtkit reads `jwt.secret`). Most kits self-initialize via `init()` or on first use reading configkit.
5. `go build ./...` and `go vet ./...` — must be green.

## Steps to DISABLE a kit

1. Remove all imports and usages of the kit from business code (`mod/*/`).
2. Optionally delete `pkg/service/<kit>/` directory if user wants it gone physically.
3. Remove its config keys from `pkg/cli/configkey/<kit>.go` if no longer referenced.
4. Remove its dependency from `go.mod` via `go mod tidy`.
5. `go build ./...` — must still compile.

## Business module wiring (mod/<name>)

Business modules are NOT optional infrastructure — they are enabled by default once created via `scaffold-generate`. To toggle a whole business module:

### Enable a business module

In `main.go`:
```go
import (
    "github.com/example/go-ai-scaffold/mod/<name>"
    "github.com/example/go-ai-scaffold/pkg/service/restkit"
)

func main() {
    cli.RootCMD(&cobra.Command{
        Use: "main",
        Run: func(cmd *cobra.Command, args []string) {
            restkit.AddActions(user.All()...)
            restkit.AddActions(<name>.All()...)  // 新增
            _ = restkit.Run()
        },
    })
    // ...
}
```

### Disable a business module

Remove its `restkit.AddActions(<name>.All()...)` line and its import from `main.go`. The `mod/<name>/` directory can stay (won't compile into binary if not imported) or be deleted.

## Rules

- Always verify build after wiring changes: `go build ./...` + `go vet ./...`.
- Never leave dangling imports.
- If a kit is missing, list available kits (from the table above) and ask which to create.
- Keep `main.go` changes minimal and grouped by module.
- Do NOT create an `internal/modules/` directory — the scaffold does not use that layout. Optional kits go in `pkg/service/`, business modules in `mod/`.
- Config keys live in `pkg/cli/configkey/`, one file per concern (redis.go, jwt.go, db.go, etc.), as `const` strings.
