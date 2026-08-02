package user

import (
	"github.com/example/go-ai-scaffold/mod/user/action/role"
	"github.com/example/go-ai-scaffold/mod/user/action/smscode"
	"github.com/example/go-ai-scaffold/mod/user/action/user"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

// All 用户、部门、角色模块
func All() []func(r *router.Router) {
	return []func(r *router.Router){user.Init, role.Init, smscode.Init}
}
