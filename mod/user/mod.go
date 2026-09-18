package user

import (
	"github.com/example/go-frame/mod/user/controller/department"
	"github.com/example/go-frame/mod/user/controller/role"
	"github.com/example/go-frame/mod/user/controller/smscode"
	"github.com/example/go-frame/mod/user/controller/user"
	"github.com/example/go-frame/pkg/service/restkit/router"
)

// All 用户、部门、角色模块
func All() []func(r *router.Router) {
	return []func(r *router.Router){user.Init, role.Init, department.Init, smscode.Init}
}
