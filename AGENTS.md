# AGENTS.md

Go scaffold for vibe-coding REST services (gin + viper + cobra + sqlx/squirrel + PostgreSQL). Repo docs/comments are in Chinese; match that style.

## Must read first

`.trae/skills/scaffold-{generate,rename,configure}/SKILL.md` — codified workflows for creating CRUD modules, renaming the project, and toggling service kits.

## Current state (verified)

- **Build is currently broken** (pre-existing): `main.go:28` calls `cmd.FrontDaoCMDNext(...)` which does not exist. Don't chase this as your own bug; if a task requires compiling, fix or restore `main.go` first.
- No tests, no lint config, no CI. The required verification for every change is:
  ```
  go build ./... && go vet ./... && gofmt -l .
  ```
- Module path `github.com/example/go-ai-scaffold` and project name `go-ai-scaffold` are placeholders; rename via `scaffold-rename` skill before use.
- Only one business module: `mod/user/`.

## Layout

```
pkg/                          # reusable infra, MUST NOT depend on mod/*
  class/                        nullable DB wrappers: String, Int64, Time, Decimal, ArrInt, MapString, File, etc.
  library/*kit/                 pure utility packages
  service/*kit/                 infra kits: restkit, sqlkit, configkit, logkit, jwtkit, rediskit, aikit, mqttkit, netkit
  cli/                          cobra root command + viper config binding
mod/<name>/                   # business module, strict 4 layers
  mod.go                        All() []func(*router.Router) — aggregate each resource's Init
  model/                        entity structs (db/json/pk/table tags only, no HTTP tags)
  dao/<resource>dao/            embeds sqlkit.Dao[T] + CascadeOpts + query methods
  service/                      function-style (no structs), business logic + TxArea orchestration
  controller/<resource>/        index.go (routing + OpenAPI), *_controller.go (handlers)
cmd/                          # extra cobra subcommands, registered via cli.AddChildCMD(...)
main.go                       # cli.RootCMD(...) → restkit.AddActions(user.All()...) → restkit.Run()
```

## Config

- Every key is a `const` in `pkg/cli/configkey/*.go`, bound as cobra flag in `pkg/cli/bind.go`, read via `configkit.GetString/GetInt/GetBool(key, default...)` — never read viper directly.
- Defaults: `:10000` for REST server; `/v3/api-docs` for swagger; `/doc.html` for knife4j UI.
- Config file: `config.yaml` in working dir, override with `-c/--config`.

## Conventions

### Layer boundaries
```
controller: ctx.BindForm(&params) → call service → ctx.JsonSuccess(ret)
service:    business logic, TxArea, cross-dao assembly, panic(exception.New("...")) on error
dao:        SQL only — parameterized Where("col=?", v), WhereJsonbPathEq, WithRecursiveRaw. Never fmt.Sprintf.
model:      struct tags: db, json, pk, table, auto, logicDel, comment, default, validate.
```

**Forbidden**: controller→dao direct call, dao business logic, service touching `*context.Context`, model with HTTP tags, `pkg/*` importing `mod/*`.

### Naming

| Element | Rule | Example |
|---|---|---|
| Module dir | `mod/<singular_lower>/` | `mod/user/` |
| Resource dir | `dao/<resource>dao/`, `controller/<resource>/` | `dao/userdao/`, `controller/user/` |
| DAO struct | `type Dao struct { sqlkit.Dao[model.User] }` | embedded |
| DAO New | `userdao.New(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao` | |
| Service | `package service`, function-style (no struct) | `func Login(username, phone, pwd string)` |
| Controller handler | `func Xxx(ctx *context.Context)` | `LoginByUsername`, `ListUsers` |
| Init | `func Init(router *router.Router)` — one per resource sub-package | in `index.go` |
| Route prefix | `/<resource>` or `/<resource>/admin` | `/user`, `/user/admin` |

### DAO cascade

```go
type CascadeOpts struct { Role bool; Department bool }
var OptsDefault = CascadeOpts{Role: true, Department: true}
// New registers cascades via dao.WithCascadeOpts(opts, func(obj *T, ctx sqlkit.CascadeCtx) { ... })
```
Batch strategy with `WithCascadeBatchLinks` to avoid N+1. No byte enums.

### Routing + OpenAPI

```go
router.Group("/user/login").Post("", Login).Api(
    openapi.Tag("user:用户模块"),
    openapi.Summary("登录"),
    openapi.ReqParam(loginParam{}),
    openapi.Response(ResLogin{}),
    openapi.Security(nil),  // nil = no auth required
)
```
- Mandatory: `Tag`, `Summary`.
- Guard with `middleware.AuthJWT()`; open endpoints get `openapi.Security(nil)`.
- All auth endpoints check JWT (stores uid only); logout via `ctx.DestroyJwt()`.

### Response format
```json
{"result": 0, "message": "", "data": {}, "total": 0}
```
- 0 = success, 401 = auth failure, 500 = business error (from `panic(exception.New(...))`, caught by `middleware.Recover`).
- `ctx.JsonSuccess(data...)`, `ctx.JsonSuccessWithPage(data, total)`, `ctx.JsonError(msg)`.

### Sensitive fields / security
- `Pwd` and similar must be `json:"-"`.
- Codebase still uses MD5 for passwords (`cryptokit.MD5`); new code must use bcrypt/scrypt/argon2.
- Never hardcode secrets/connection strings.
- Delete operations must nullify unique fields (phone, username) to avoid dirty data.

## Adding a module

Use `scaffold-generate` skill. Wire the controller's `Init` into `mod.go` `All()` and register in `main.go`. Verify: `go build ./... && go vet ./... && gofmt -l .`.