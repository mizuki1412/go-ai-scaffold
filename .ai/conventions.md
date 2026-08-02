# AI Conventions — go-ai-scaffold

> AI 在对本工程做任何改动前，请先读完本文件。它是脚手架的「宪法」。
> 本文档由 `mod/user` 模块的真实代码归纳而来，所有规则以 `mod/user` 为唯一参考实现。

## 1. 占位符

- **Module 路径占位符**: `github.com/example/go-ai-scaffold`
- **工程名占位符**: `go-ai-scaffold`

复制本脚手架后，第一步必须用 `scaffold-rename` skill 把这两个值改成新项目的真实值。改完后整工程必须仍能 `go build ./...` 通过。

## 2. 分层架构（严格遵循）

```
HTTP 请求
  └─ controller   入参绑定（BindForm）/ 调 service / 用 ctx.JsonSuccess 统一响应
       └─ service     业务逻辑 / 事务编排 / 跨 dao 数据装配 / 参数校验（panic 抛 exception）
            └─ dao        仅做 DB 访问，sqlkit.Dao 泛型基类 + CascadeOpts 级联策略
                 └─ model       实体定义，带 db/json tag，不写 HTTP 标签
pkg/*            可被任何项目复用的纯工具库（class/library/cli/service）
mod/<name>       业务模块，每个模块内含 model/dao/service/controller 四层
```

**禁止**：
- controller 直接调 dao（必须经 service）
- dao 写业务逻辑（dao 只做 SQL 构造与执行）
- service 直接操作 `*context.Context`（service 不感知 HTTP 层）
- 在 model 里写 HTTP 标签（仅 `db`/`json`/`pk`/`table`/`auto`/`logicDel`/`comment`/`default`/`validate`）
- 在 pkg/* 里反向依赖 mod/*（pkg 必须可独立复用）

## 3. 模块目录结构

每个业务模块位于 `mod/<name>/`，内部固定四层：

```
mod/user/
  mod.go                          // 模块入口，暴露 All() []func(*router.Router)
  model/
    user.go                       // 实体 struct + 自定义 List 类型
    role.go
  dao/
    userdao/
      userdao.go                  // Dao struct（嵌入 sqlkit.Dao[T]）+ CascadeOpts + New + 查询方法
    roledao/
      roledao.go
  service/
    user_service.go               // 包 service，业务函数（非 struct，函数式风格）
    role_service.go
    department_service.go
  controller/
    user/
      index.go                    // Init(router) 注册路由 + openapi 元数据
      user_controller.go          // 用户自助接口
      user_admin_controller.go    // 管理员接口
    role/
      index.go
      role_controller.go
    department/                   // department 单独成包，URL 前缀 /department/*
      department_controller.go    // 含 Init，路由+控制器同文件
    smscode/
      sms_code_controller.go
```

## 4. 命名约定

| 元素 | 规则 | 例 |
|---|---|---|
| 模块名 | 单数小写 | `user`, `role` |
| 资源名 | 单数 PascalCase | `User`, `Role` |
| model 文件 | `mod/<name>/model/<snake>.go` | `user.go` |
| model 结构体 | 资源名 | `User` |
| model 自定义列表 | `<Resource>List` | `UserList` |
| dao 目录 | `mod/<name>/dao/<resource>dao/` | `dao/userdao/` |
| dao 文件 | `mod/<name>/dao/<resource>dao/<resource>dao.go` | `dao/userdao/userdao.go` |
| dao 结构体 | `Dao`（嵌入 `sqlkit.Dao[T]`） | `type Dao struct { sqlkit.Dao[model.User] }` |
| dao 构造 | `New(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao` | `userdao.New(userdao.OptsDefault)` |
| service 文件 | `mod/<name>/service/<resource>_service.go` | `user_service.go` |
| service 包 | `package service`（函数式，无 struct） | `func Login(...)` |
| controller 目录 | `mod/<name>/controller/<resource>/` | `controller/user/` |
| controller 文件 | `mod/<name>/controller/<resource>/<resource>_controller.go` | `user_controller.go` |
| controller 函数 | 大驼峰，动词开头 | `Login`, `ListUsers`, `CreateRole` |
| 路由注册函数 | `Init(router *router.Router)` | 每个资源子包一个 Init |
| 路由前缀 | `/<resource>` 或 `/<resource>/admin` | `/user`, `/user/admin` |
| 请求参数结构 | `<action>Param` 或 `<action>Params` | `loginParam`, `listUsersParams` |

### 4.1 特例：部门单独成包

部门（department）控制器必须单独成包 `controller/department/`，URL 前缀 `/department/*`，避免与 `controller/user/` 下的用户接口混淆。文件内同时含 `Init` 与控制器函数。

## 5. DAO 层约定

### 5.1 CascadeOpts 替代 byte 枚举

每个 dao 必须定义自己的 `CascadeOpts struct`，字段为 `bool`，描述级联策略：

```go
type CascadeOpts struct {
    Role       bool
    Department bool
}
var (
    OptsNone     = CascadeOpts{}
    OptsDefault  = CascadeOpts{Role: true, Department: true}
    OptsRoleOnly = CascadeOpts{Role: true}
)
```

构造函数通过 `WithCascadeOpts` 注册级联回调：

```go
func New(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao {
    dao := sqlkit.New[model.User](ds...)
    dao = dao.WithCascadeOpts(opts, func(obj *model.User, ctx sqlkit.CascadeCtx) {
        o := ctx.Opts.(CascadeOpts)
        if o.Role && obj.Role != nil {
            obj.Role = roledao.New(roledao.OptsDefault, ctx.Ds).SelectOneWithDelById(obj.Role.Id)
        }
    })
    return Dao{dao}
}
```

**禁止**：使用旧的 byte 枚举方式（已废弃）。

### 5.2 SQL 安全

- JSONB 查询必须用 `WhereJsonbPathEq`，禁止 `fmt.Sprintf` 拼接
- 递归 CTE 必须用 `WithRecursiveRaw`，禁止字符串拼接 CTE body
- 参数化查询用 `Where("col=?", val)`，禁止 `fmt.Sprintf` 拼 WHERE

### 5.3 事务

跨表操作必须用 `sqlkit.TxArea(func(ds *sqlkit.DataSource) { ... })`，dao 通过 `New(opts, ds)` 继承事务连接。

### 5.4 数据源

接口层不传 schema/dataSource，DAO 内部用 `sqlkit.DefaultDataSource()`。多租户场景通过 `sqlkit.DefaultDataSource().WithSchema()` 或 dao 的 `ds` 参数注入。

## 6. Service 层约定

- `package service`，函数式风格（不定义 service struct）
- 业务错误统一 `panic(exception.New("..."))`，由中间件 Recover 兜底转 HTTP 响应
- 参数结构定义在 service 文件中，controller 用类型别名引用：`type listUsersParams = service.ListUsersParams`
- 跨 dao 数据装配在 service 完成，controller 不直接调 dao
- 事务编排用 `sqlkit.TxArea`

## 7. Controller 层约定

- 薄层：仅做参数绑定（`ctx.BindForm(&params)`）→ 调 service → `ctx.JsonSuccess(ret)`
- 函数签名统一 `func Xxx(ctx *context.Context)`
- 请求参数结构定义在 controller 文件，或用 `type xxxParams = service.XxxParams` 引用
- 路由 + openapi 元数据在 `index.go` 中通过链式 `.Api(...)` 注册

### 7.1 路由注册示例

```go
func Init(router *router.Router) {
    tag := "user:用户模块"
    // 无需鉴权的接口
    router.Group("/user/login").Post("", Login).Api(openapi.Tag(tag),
        openapi.Summary("登录"),
        openapi.ReqParam(loginParam{}),
        openapi.Response(ResLogin{}),
        openapi.Security(nil))  // nil 表示该接口不鉴权
    // 需鉴权的接口
    r := router.Group("/user", middleware.AuthJWT())
    {
        r.Get("/logout", Logout).Api(openapi.Tag(tag), openapi.Summary("登出"))
        r.Post("/updatePwd", UpdatePwd).Api(openapi.Tag(tag),
            openapi.Summary("密码修改"),
            openapi.ReqParam(updatePwdParam{}))
    }
}
```

### 7.2 模块入口

每个模块根目录 `mod.go` 暴露 `All()` 聚合所有资源的 Init：

```go
package user
func All() []func(r *router.Router) {
    return []func(r *router.Router){user.Init, role.Init, department.Init, smscode.Init}
}
```

## 8. 响应格式

所有 controller 必须用 `ctx.JsonSuccess(data...)` 返回，统一结构：

```json
{
  "result": 0,
  "message": "ok",
  "data": {},
  "total": 0
}
```

- `result`: 0 成功，401 鉴权失败，500 业务错误
- `data`: 业务数据，可选
- `total`: 列表分页时的总数，可选

错误响应由 `middleware/recover.go` 统一兜底，controller/service 用 `panic(exception.New(...))` 抛出。

## 9. 鉴权

- `middleware.AuthJWT()` 中间件校验 JWT
- 登录接口用 `openapi.Security(nil)` 标记免鉴权
- JWT 仅存储 uid，多租户/schema 隔离通过 sqlkit 处理
- 注销用 `ctx.DestroyJwt()`，缓存 key 用登录时签发的原始 token 字符串

## 10. OpenAPI 文档

- 每个 `.Post/.Get/.Put/.Delete` 后链式 `.Api(...)` 声明元数据
- `openapi.Tag("name:描述")` 必填，用于分组
- `openapi.Summary("...")` 必填
- `openapi.ReqParam(struct{})` 声明 query/path 参数
- `openapi.ReqBody(struct{})` 声明请求体
- `openapi.Response(struct{})` 或 `openapi.Response([]*T{})` 声明响应
- `openapi.Security(nil)` 标记免鉴权接口

## 11. 新增模块的标准流程

调用 `scaffold-generate` skill，它会模仿 `mod/user` 的结构生成 model/dao/service/controller 四层文件并接入 `mod.go`。

## 12. 构建验证

任何改动后必须通过：

```bash
go build ./...
go vet ./...
gofmt -l .
```

skill 在结束时必须跑这三条，不绿不算完。

## 13. 不要做的事

- 不要给 pkg/* 里的工具加 mod/* 的业务依赖
- 不要让 model 直接出现在 JSON 响应里而不经过 service 处理（敏感字段如 Pwd 必须用 `json:"-"`）
- 不要新建随意命名的包；新资源一律按第 3 节结构放进对应层
- 不要在脚手架里硬编码业务密钥/连接串
- 不要用 MD5 做密码哈希（现有代码待迁移，新代码必须用 bcrypt/scrypt/argon2）
- 不要复用已删除的 user 记录（脏数据风险），删除时必须清理 phone/username 等唯一字段
- 不要在 DAO 中拼接 SQL 字符串，必须用 sqlkit 的参数化方法
