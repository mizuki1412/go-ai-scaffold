package service

import (
	"time"

	"github.com/example/go-ai-scaffold/mod/user/dao/departmentdao"
	"github.com/example/go-ai-scaffold/mod/user/dao/privilegedao"
	"github.com/example/go-ai-scaffold/mod/user/dao/roledao"
	"github.com/example/go-ai-scaffold/mod/user/dao/userdao"
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
)

// ============ ListAllPrivileges ============

func ListAllPrivileges() []*model.PrivilegeConstant {
	dao := privilegedao.New(privilegedao.OptsNone)
	return dao.ListPrivileges()
}

// ============ CreateRole ============

type CreateRoleParams struct {
	Name           string          `validate:"required"`
	PrivilegesJson class.ArrString `validate:"required" default:"[]" comment:"[a,b,c]"`
	DepartmentId   class.Int64
	Extend         class.MapString
}

func CreateRole(params CreateRoleParams) {
	role := &model.Role{}
	if params.DepartmentId.Valid {
		departmentDao := departmentdao.New(departmentdao.OptsNone)
		department := departmentDao.SelectOneById(params.DepartmentId)
		if department == nil {
			panic(exception.New("部门不存在"))
		}
		role.Department = department
	}
	role.Name.Set(params.Name)
	role.Privileges = params.PrivilegesJson
	role.CreateDt.Set(time.Now())
	role.Extend.Set(params.Extend)
	rdao := roledao.New(roledao.OptsDefault)
	rdao.InsertObj(role)
}

// ============ UpdateRole ============

type UpdateRoleParams struct {
	Id             int64 `validate:"required"`
	Name           class.String
	PrivilegesJson class.ArrString `comment:"数组json字符串：[a,b,c]"`
	DepartmentId   class.Int64
	Extend         class.MapString
}

func UpdateRole(params UpdateRoleParams) {
	dao := roledao.New(roledao.OptsDefault)
	role := dao.SelectOneById(params.Id)
	if role == nil {
		panic(exception.New("角色不存在"))
	}
	if params.DepartmentId.Valid && (role.Department == nil || params.DepartmentId.Int64 != role.Department.Id) {
		departmentDao := departmentdao.New(departmentdao.OptsNone)
		d := departmentDao.SelectOneById(params.DepartmentId.Int64)
		if d == nil {
			panic(exception.New("部门不存在"))
		}
		role.Department = d
	}
	if params.Name.Valid {
		role.Name.Set(params.Name.String)
	}
	if params.PrivilegesJson.Valid {
		role.Privileges = params.PrivilegesJson
	}
	if params.Extend.IsValid() {
		role.Extend.PutAll(params.Extend.Map)
	}
	dao.UpdateObj(role)
}

// ============ DeleteRole ============

type DeleteRoleParams struct {
	Id int64 `validate:"required"`
}

// DeleteRole B13: 用 GetBool 替代类型断言，避免 panic。
func DeleteRole(id int64) {
	dao := roledao.New(roledao.OptsNone)
	role := dao.SelectOneById(id)
	if role == nil {
		panic(exception.New("角色不存在"))
	}
	if role.Extend.GetBool("immutable") { // B13
		panic(exception.New("该角色不可删除"))
	}
	userDao := userdao.New(userdao.OptsNone)
	rid := class.Int64{}
	rid.Set(id)
	us := userDao.List(userdao.ListParam{RoleId: rid})
	if us != nil && len(us) > 0 {
		panic(exception.New("角色下还有用户,不能删除"))
	}
	dao.DeleteById(role.Id)
}

// ============ ListRoles ============

type ListRolesParam struct {
	Root class.Int64 `comment:"指定根department"`
}

// ListRoles B14: 用 model.EnsurePrivileges 替代 controller 中的循环修补。
func ListRoles(params ListRolesParam) []*model.Role {
	dao := roledao.New(roledao.OptsDefault)
	var roles []*model.Role
	if params.Root.Valid {
		roles = dao.ListFromRootDepart(params.Root.Int64)
	} else {
		roles = dao.List(roledao.ListParam{})
	}
	for _, r := range roles {
		r.EnsurePrivileges() // B14
	}
	return roles
}

// ============ ListRolesWithUser ============

type ListRolesWithUserParams struct {
	RoleId int64 `validate:"required"`
}

// ListRolesWithUser B12: 修复 bug — 原代码循环中每个 role 都查 params.RoleId 的用户，
// 应改为查当前 r.Id 的用户。
func ListRolesWithUser(params ListRolesWithUserParams) []*model.Role {
	dao := roledao.New(roledao.OptsDefault)
	list := dao.List(roledao.ListParam{})
	udao := userdao.New(userdao.OptsDefault)
	for _, r := range list {
		r.EnsurePrivileges()
		r.Extend.PutAll(map[string]any{
			"users": udao.List(userdao.ListParam{Roles: []int64{r.Id}}), // B12: r.Id 而非 params.RoleId
		})
	}
	return list
}
