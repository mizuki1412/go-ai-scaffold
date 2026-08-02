# AI Conventions — go-ai-scaffold

> AI 在对本工程做任何改动前，请先读完本文件。它是脚手架的「宪法」。

## 1. 占位符

- **Module 路径占位符**: `github.com/example/go-ai-scaffold`
- **工程名占位符**: `go-ai-scaffold`

复制本脚手架后，第一步必须用 `scaffold-rename` skill 把这两个值改成新项目的真实值。改完后整工程必须仍能 `go build ./...` 通过。

## 2. 分层架构（严格遵循）

```
HTTP 请求
  └─ controller   入参校验 / 调 service / 用 pkg/response 统一响应
       └─ service     业务逻辑 / 事务编排 / entity↔dto 转换
            └─ dao        仅做 DB 访问，不写业务
                 └─ model/entity  GORM 模型
model/dto        请求/响应结构体（不绑定到 DB）
pkg/*            可被任何项目复用的纯工具库
```

**禁止**：
- controller 直接调 dao
- dao 写业务逻辑
- service 直接操作 `*gin.Context`
- 在 model/entity 里写 HTTP 标签（如 `json:""` 之外的 binding）

## 3. 命名约定

| 元素 | 规则 | 例 |
|---|---|---|
| 资源名 | 单数 PascalCase | `User`, `Post` |
| entity 文件 | `internal/model/entity/<snake>.go` | `user.go` |
| entity 结构体 | 资源名 | `User` |
| dto 文件 | `internal/model/dto/<snake>.go` | `user.go` |
| dto 结构体 | `<Resource><CreateReq/UpdateReq/Query/Resp>` | `UserCreateReq` |
| dao 文件 | `internal/dao/<snake>.go` | `user.go` |
| dao 结构体 | `<Resource>DAO` | `UserDAO` |
| dao 构造 | `New<Resource>DAO(db)` | `NewUserDAO` |
| service 文件 | `internal/service/<snake>.go` | `user.go` |
| service 结构体 | `<Resource>Service` | `UserService` |
| controller 文件 | `internal/controller/<snake>.go` | `user.go` |
| controller 结构体 | `<Resource>Controller` | `UserController` |
| 路由前缀 | `/<snake_plural>` | `/users` |

## 4. 新增模块的标准流程

调用 `scaffold-generate` skill。它会模仿 `internal/**/user.go` 这套示例生成全部 5 层文件并接入 router。

## 5. 可选模块

位置：`internal/modules/<name>/`。默认不接入 router。启用走 `scaffold-configure` skill。

可选：`auth`、`cache`、`queue`、`websocket`、`upload`。

## 6. 响应格式

所有 controller 必须用 `pkg/response` 返回，统一结构：

```json
{ "code": 0, "message": "ok", "data": {} }
```

## 7. 错误处理

- service 返回 `error`，不要在 service 里直接写 HTTP 响应。
- controller 用 `pkg/response` 把 error 映射成 code/message。
- panic 由 `internal/middleware/recovery.go` 统一兜底。

## 8. 构建验证

任何改动后必须通过：

```bash
go build ./...
go vet ./...
```

skill 在结束时必须跑这两条，不绿不算完。

## 9. 不要做的事

- 不要给 pkg/* 里的工具加业务依赖。
- 不要把 model/entity 当 dto 用（不要让 entity 直接出现在 JSON 响应里）。
- 不要新建随意命名的包；新资源一律放进现有 5 层包。
- 不要在脚手架里硬编码业务密钥/连接串。
