---
name: "scaffold-generate"
description: "Generate a full CRUD resource module (model/dao/service/controller + route wiring) by mimicking the example mod/user module. Invoke when user says '生成 post 模块' / 'add a Product resource' / '生成文章的 CRUD'."
---

# Scaffold Generate

Generate a new resource module by COPYING THE PATTERN of the existing `mod/user` module. The `mod/user` module is the single source of truth for conventions — follow it exactly.

## When to invoke

- User asks to generate a resource: "生成 post 模块" / "add Product CRUD" / "做一个订单的增删改查".
- User asks to scaffold an entity.

## Inputs

Ask the user (only if not obvious) for:
1. `module` — module name, lowercase singular, e.g. `post`, `order`. If user says "生成文章模块", module = `article`.
2. `resource` — singular, PascalCase, e.g. `Post`, `Order`. Default: UpperFirst(module).
3. `table` — DB table name (default: `sys_<module>` or per user spec).
4. `fields` — list of `{name, go_type, json_tag, db_tag, comment}`. If user gave a natural-language description, infer types:
   - string → `class.String`
   - int → `class.Int32` 或 `class.Int64`
   - long text → `class.String`
   - bool → `class.Bool`
   - time → `class.Time`
   - decimal/money → `class.Decimal`
   - string array → `class.ArrString`
   - ext map → `class.MapString`

## Reference module to mimic

Read these files first and mirror their structure, naming, comments style EXACTLY:
- `mod/user/model/user.go` — entity struct with `db`/`json`/`pk`/`table`/`auto`/`logicDel` tags + custom `XxxList` type
- `mod/user/dao/userdao/userdao.go` — `Dao` embedding `sqlkit.Dao[T]` + `CascadeOpts` + `New(opts, ds...)` + query methods
- `mod/user/service/user_service.go` — `package service` functional style, `XxxParams` structs, business logic
- `mod/user/controller/user/index.go` — `Init(router)` registering routes with `.Api(openapi.Tag/Summary/ReqParam/ReqBody/Response/Security)`
- `mod/user/controller/user/user_controller.go` — thin handlers: `BindForm` → call service → `JsonSuccess`
- `mod/user/mod.go` — `All()` aggregating all resources' `Init`

## Steps

### 1. Model — `mod/<module>/model/<resource_snake>.go`

```go
package model

import (
    "database/sql/driver"
    "github.com/example/go-ai-scaffold/pkg/class"
    "github.com/spf13/cast"
)

type <Resource> struct {
    Id       int64           `json:"id,omitempty" db:"id" pk:"true" table:"sys_<module>" auto:"true"`
    Name     class.String    `json:"name,omitempty" db:"name"`
    // ...fields from input...
    Deleted  class.Bool      `json:"-" db:"deleted" logicDel:"true"`
    Extend   class.MapString `json:"extend,omitempty" db:"extend"`
    CreateDt class.Time      `json:"createDt,omitempty" db:"createdt"`
}

func (th *<Resource>) Scan(value any) error {
    if value == nil { return nil }
    th.Id = cast.ToInt64(value)
    return nil
}
func (th *<Resource>) Value() (driver.Value, error) { return th.Id, nil }

type <Resource>List []*<Resource>
func (l <Resource>List) Len() int           { return len(l) }
func (l <Resource>List) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l <Resource>List) Less(i, j int) bool { return l[i].Id < l[j].Id }
```

- 敏感字段（如 Pwd）必须 `json:"-"`
- `Deleted` 字段用 `class.Bool` + `logicDel:"true"` 标记逻辑删除
- 表名通过 `table:"sys_xxx"` tag 声明，不写 `TableName()` 方法

### 2. DAO — `mod/<module>/dao/<resource>dao/<resource>dao.go`

```go
package <resource>dao

import (
    "github.com/example/go-ai-scaffold/mod/<module>/model"
    "github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

type Dao struct {
    sqlkit.Dao[model.<Resource>]
}

// CascadeOpts 级联策略，替代旧的 byte 枚举
type CascadeOpts struct {
    // 按需定义级联字段，如：
    // User bool
    // Role bool
}

var (
    OptsNone    = CascadeOpts{}
    OptsDefault = CascadeOpts{} // 按需填充默认级联
)

func New(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao {
    dao := sqlkit.New[model.<Resource>](ds...)
    dao = dao.WithCascadeOpts(opts, func(obj *model.<Resource>, ctx sqlkit.CascadeCtx) {
        o := ctx.Opts.(CascadeOpts)
        _ = o // 按需级联
    })
    return Dao{dao}
}

// 按需添加查询方法
func (dao Dao) FindByName(name string) *model.<Resource> {
    return dao.Select().Where("name=?", name).One()
}

type ListParam struct {
    IdList []int64
    Page   *sqlkit.Page
}

func (dao Dao) List(param ListParam) model.<Resource>List {
    builder := dao.Select().OrderBy("id")
    if len(param.IdList) > 0 {
        builder = builder.WhereUnnestIn("id", param.IdList)
    }
    if param.Page != nil && param.Page.PageSize > 0 {
        list, _ := builder.Page(*param.Page)
        return list
    }
    return builder.List()
}
```

**关键**：
- JSONB 查询用 `WhereJsonbPathEq`
- 递归 CTE 用 `WithRecursiveRaw`
- 禁止 `fmt.Sprintf` 拼 SQL

### 3. Service — `mod/<module>/service/<resource>_service.go`

```go
package service

import (
    "time"
    "github.com/example/go-ai-scaffold/mod/<module>/dao/<resource>dao"
    "github.com/example/go-ai-scaffold/mod/<module>/model"
    "github.com/example/go-ai-scaffold/pkg/class"
    "github.com/example/go-ai-scaffold/pkg/class/exception"
    "github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

// ============ Create ============

type Create<Resource>Params struct {
    Name class.String `validate:"required"`
    // ...
}

func Create<Resource>(params Create<Resource>Params) *model.<Resource> {
    dao := <resource>dao.New(<resource>dao.OptsNone)
    obj := &model.<Resource>{}
    if params.Name.Valid {
        obj.Name.Set(params.Name.String)
    }
    obj.CreateDt.Set(time.Now())
    dao.InsertObj(obj)
    return obj
}

// ============ Update ============

type Update<Resource>Params struct {
    Id   int64 `validate:"required"`
    Name class.String
}

func Update<Resource>(params Update<Resource>Params) {
    dao := <resource>dao.New(<resource>dao.OptsDefault)
    obj := dao.SelectOneById(params.Id)
    if obj == nil {
        panic(exception.New("<resource>不存在"))
    }
    if params.Name.Valid {
        obj.Name.Set(params.Name.String)
    }
    dao.UpdateObj(obj)
}

// ============ Delete ============

type Delete<Resource>Params struct {
    Id int64 `validate:"required"`
}

func Delete<Resource>(id int64) {
    dao := <resource>dao.New(<resource>dao.OptsNone)
    obj := dao.SelectOneById(id)
    if obj == nil {
        panic(exception.New("<resource>不存在"))
    }
    dao.DeleteById(id)
}

// ============ List ============

type List<Resource>Param struct {
    Page *sqlkit.Page
}

func List<Resource>s(params List<Resource>Param) []*model.<Resource> {
    dao := <resource>dao.New(<resource>dao.OptsDefault)
    return dao.List(<resource>dao.ListParam{Page: params.Page})
}
```

**关键**：
- `package service`，函数式风格（无 struct）
- 业务错误 `panic(exception.New("..."))`
- 跨表操作用 `sqlkit.TxArea`
- 参数结构定义在此，controller 用 `type xxxParams = service.XxxParams` 引用

### 4. Controller — `mod/<module>/controller/<resource>/`

#### 4.1 `index.go` — 路由注册

```go
package <resource>

import (
    "github.com/example/go-ai-scaffold/mod/<module>/model"
    "github.com/example/go-ai-scaffold/pkg/service/restkit/middleware"
    "github.com/example/go-ai-scaffold/pkg/service/restkit/openapi"
    "github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

func Init(router *router.Router) {
    tag := "<module>:<中文描述>"
    r := router.Group("/<resource>").Use(middleware.AuthJWT())
    r.Post("/list", List).Api(openapi.Tag(tag),
        openapi.Summary("<resource>列表"),
        openapi.ReqBody(listParams{}),
        openapi.Response([]*model.<Resource>{}))
    r.Post("/create", Create).Api(openapi.Tag(tag),
        openapi.Summary("<resource>新增"),
        openapi.ReqBody(createParams{}))
    r.Post("/update", Update).Api(openapi.Tag(tag),
        openapi.Summary("<resource>修改"),
        openapi.ReqBody(updateParams{}))
    r.Get("/del", Delete).Api(openapi.Tag(tag),
        openapi.Summary("<resource>删除"),
        openapi.ReqParam(deleteParams{}))
}
```

#### 4.2 `<resource>_controller.go` — 控制器

```go
package <resource>

import (
    "github.com/example/go-ai-scaffold/mod/<module>/service"
    "github.com/example/go-ai-scaffold/pkg/service/restkit/context"
)

type createParams = service.Create<Resource>Params

func Create(ctx *context.Context) {
    params := createParams{}
    ctx.BindForm(&params)
    obj := service.Create<Resource>(params)
    ctx.JsonSuccess(obj)
}

type updateParams = service.Update<Resource>Params

func Update(ctx *context.Context) {
    params := updateParams{}
    ctx.BindForm(&params)
    service.Update<Resource>(params)
    ctx.JsonSuccess()
}

type deleteParams = service.Delete<Resource>Params

func Delete(ctx *context.Context) {
    params := deleteParams{}
    ctx.BindForm(&params)
    service.Delete<Resource>(params.Id)
    ctx.JsonSuccess()
}

type listParams = service.List<Resource>Param

func List(ctx *context.Context) {
    params := listParams{}
    ctx.BindForm(&params)
    ctx.JsonSuccess(service.List<Resource>s(params))
}
```

**关键**：
- 薄层：`BindForm` → 调 service → `JsonSuccess`
- 函数签名统一 `func Xxx(ctx *context.Context)`
- 参数结构用 `type xxxParams = service.XxxParams` 别名引用

### 5. Wire into `mod.go`

```go
package <module>

import (
    "github.com/example/go-ai-scaffold/mod/<module>/controller/<resource>"
    "github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

func All() []func(r *router.Router) {
    return []func(r *router.Router){<resource>.Init}
}
```

如模块已有其他资源，把 `<resource>.Init` 追加到现有 `All()` 返回切片。

### 6. Wire module into app entry

在应用入口（通常是 `main.go` 或 `cmd/.../main.go`）调用 `mod.<module>.All()`，将返回的 Init 函数列表注册到 router。

### 7. Verify

```bash
go build ./...
go vet ./...
gofmt -l .
```

必须全绿。skill 结束前必须跑这三条。

## Naming conventions

| Layer | File | Struct/Func | Constructor |
|-------|------|-------------|-------------|
| model | `mod/<m>/model/<snake>.go` | `<Resource>` | — |
| dao | `mod/<m>/dao/<resource>dao/<resource>dao.go` | `Dao` | `New(opts, ds...) Dao` |
| service | `mod/<m>/service/<resource>_service.go` | `package service` 函数 | — |
| controller | `mod/<m>/controller/<resource>/<resource>_controller.go` | `func Xxx(ctx)` | — |
| router | `mod/<m>/controller/<resource>/index.go` | `Init(router)` | — |

## Rules

- Always mimic the `mod/user` example exactly — same package layout, same comment style (`// ============ Xxx ============` 分段), same response helpers.
- Never invent new packages; everything goes into the existing layer packages (`model`, `dao/<resource>dao`, `service`, `controller/<resource>`).
- Always wire the new controller's `Init` into `mod.go`'s `All()` — an unwired controller is a bug.
- Always run `go build ./...` at the end; do not leave the project broken.
- If a field type is ambiguous, ask the user — do not guess silently.
- Use `class.*` types (class.String, class.Int64, class.Time, etc.) for nullable DB columns, not `*string` / `sql.NullString`.
- Sensitive fields must use `json:"-"`.
