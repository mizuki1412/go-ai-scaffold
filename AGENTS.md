# AGENTS.md

Go scaffold for vibe-coding REST services (gin + viper + cobra + sqlx/squirrel + PostgreSQL). Repo docs/comments are in Chinese; match that style.

## Must read first

- `.ai/conventions.md` — the authoritative, detailed convention doc ("宪法"). Read it fully before modifying any business code. `mod/user` is the single reference implementation.
- `.trae/skills/scaffold-{generate,rename,configure}/SKILL.md` — codified workflows for creating CRUD modules, renaming the project, and toggling service kits. Follow them even though they live under `.trae/`.

## Current state (verified)

- Build is currently broken and pre-existing: `main.go:28` calls `cmd.FrontDaoCMDNext(...)`, but its defining file `cmd/gen_front_dao.go` (plus 3 other `cmd/*.go` files) are staged for deletion. Don't chase this as your own bug; if a task requires compiling, fix or restore `main.go`/those files first.
- No tests, no lint config, no CI in the repo. The required verification for every change is:
  ```
  go build ./... ; go vet ./... ; gofmt -l .
  ```
- `github.com/example/go-ai-scaffold` (module path) and `go-ai-scaffold` (project name) are placeholders. Copying the scaffold to a new project requires a full rename (see `scaffold-rename` skill).

## Layout

- `pkg/` — reusable infrastructure, must NOT depend on `mod/*`:
  - `pkg/class/` base types, `pkg/library/*kit` pure utilities, `pkg/cli/` cobra+config wiring.
  - `pkg/service/*kit` — infra kits: `restkit` (HTTP/router/middleware/openapi/context), `sqlkit` (generic DAO), `rediskit`, `jwtkit`, `aikit`, `mqttkit`, `logkit`, `configkit`, etc.
- `mod/<name>/` — one business module, strict 4 layers: `model/`, `dao/<resource>dao/`, `service/` (function-style, no structs), `controller/<resource>/`. `mod.go` exposes `All() []func(*router.Router)` aggregating each resource's `Init(router)`.
- `cmd/` — extra cobra subcommands registered via `cli.AddChildCMD(...)`.
- `main.go` — wires modules: `restkit.AddActions(user.All()...)` then `restkit.Run()`.

## Config system

- Every config key is a `const` in `pkg/cli/configkey/*.go`, bound as a cobra flag in `pkg/cli/bind.go` and read via viper. Values come from CLI flags or `config.yaml` in the working dir (override with `-c/--config`). `pkg/cli/configkey/` is the source of truth for key names (e.g. port is `rest.port`, default 10000).
- At runtime read config with `configkit.GetString/GetInt/GetBool(key, default...)` — never read viper directly.
- Default REST server runs on `:10000`; swagger at `/v3/api-docs`, knife4j UI at `/doc.html`.

## Hard rules (see conventions.md for details)

- Layers: controller binds params (`ctx.BindForm`) → calls service → `ctx.JsonSuccess(...)`. controller must never call dao; service must never touch `*context.Context`.
- Errors are `panic(exception.New("..."))`; `middleware.Recover` converts them to `{result:500,...}`.
- Auth: guard routes with `middleware.AuthJWT()`, mark open endpoints with `openapi.Security(nil)`.
- Nullable/DB columns use `class.*` wrapper types (`class.String`, `class.Int64`, `class.Time`, `class.Decimal`, ...), never `*string`/`sql.NullString`. Sensitive fields (e.g. `Pwd`) must be `json:"-"`.
- SQL safety: parameterized `Where("col=?", v)`, JSONB via `WhereJsonbPathEq`, recursive CTE via `WithRecursiveRaw`. Never build SQL with `fmt.Sprintf`.
- DAO cascade strategy via per-dao `CascadeOpts` struct + `WithCascadeBatchLinks` (batch to avoid N+1); not byte enums.
- Transactions: `sqlkit.TxArea(func(ds *sqlkit.DataSource){...})`; daos inherit the txn connection via `New(opts, ds)`.
- Routing + OpenAPI metadata are chained `.Api(openapi.Tag(...), openapi.Summary(...), ...)` after `.Get/.Post/.Put/.Delete/.GetPost` in each resource's `index.go`.

## Adding a resource/module

Use the `scaffold-generate` skill, which mirrors `mod/user`. Always wire the new controller's `Init` into `mod.go` `All()` and register the module in `main.go`; an unwired controller is a bug. Finish with the verification trio.
