# AGENTS.md

Go scaffold for vibe-coding REST services (gin + viper + cobra + sqlx/squirrel + PostgreSQL). Repo docs/comments are in Chinese; match that style.

## Must read first

`doc/effective_go.md` — 编写本项目代码的规范依据（三源融合知识图谱：官方 Effective Go + 《Go专家编程》+ 《Mastering Go》，基准 Go 1.25，末章含本项目对照审计清单）。

## Current state (verified)

- Build, vet, gofmt and tests all green (verified 2026-08-22, after the audit round that closed 27 findings). The required verification for every change is:
  ```
  go build ./... && go vet ./... && gofmt -l . && go test ./...
  ```
- Tests cover pkg infra only (4 files: `pkg/class`, `pkg/library/cmdkit`, `pkg/library/cryptokit`, `pkg/library/framekit`); no `mod/` tests, no lint config, no CI.
- Module path `github.com/example/go-ai-scaffold` and project name `go-ai-scaffold` are placeholders; rename (go.mod module path + import prefixes, at minimum) before reuse.
- Only one business module: `mod/user/`.

## Layout

```
pkg/                          # reusable infra, MUST NOT depend on mod/*
  class/                        nullable DB wrappers: String, Int64, Time, Decimal, ArrInt, MapString, File, etc.
  library/*kit/                 pure utility packages
  service/*kit/                 infra kits: restkit, sqlkit, configkit, logkit, jwtkit, rediskit, aikit, mqttkit, netkit, cachekit, cronkit, excelkit, pdfkit, serialkit, storagekit
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
- Session model (sliding): a server-side whitelist key `token:<raw-token>` gates every authenticated request. `jwt.idle` (default 1h) is the idle window — `AuthJWT` renews it via `cachekit.Renew` on each authenticated request; `jwt.expire` (default 168h) is the absolute cap baked into the JWT exp. Set `jwt.idle<=0` to disable sliding (TTL = expire, legacy behavior).
- Token domain isolation: guard admin-domain routes with `middleware.AuthJWTScope("admin")` and issue those tokens via `jwtkit.New(uid, "admin")` — app and admin tokens share one secret/whitelist, so `Claims.Scope` is what keeps them from crossing over.

### Response format
```json
{"result": 0, "message": "", "data": {}, "total": 0}
```
- 0 = success, 401 = auth failure, 500 = business error (from `panic(exception.New(...))`, caught by `middleware.Recover`).
- `ctx.JsonSuccess(data...)`, `ctx.JsonSuccessWithPage(data, total)`, `ctx.JsonError(msg)`.

### Sensitive fields / security
- `Pwd` and similar must be `json:"-"`.
- Passwords use bcrypt (`cryptokit.HashPwd`/`CheckPwd`); legacy MD5 hashes are lazily upgraded to bcrypt on successful login (`cryptokit.NeedUpgrade`). Never write new MD5 password code.
- Never hardcode secrets/connection strings.
- Delete operations must nullify unique fields (phone, username) to avoid dirty data.

## Adding a module

Follow the Layout & Conventions above (model → dao → service → controller). Wire the controller's `Init` into `mod.go` `All()` and register in `main.go`. Verify: `go build ./... && go vet ./... && gofmt -l . && go test ./...`.