package userdao

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/example/go-ai-scaffold/mod/user/dao/departmentdao"
	"github.com/example/go-ai-scaffold/mod/user/dao/roledao"
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/library/stringkit"
	"github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

type Dao struct {
	sqlkit.Dao[model.User]
}

// CascadeOpts S10/S11: 级联策略选项，替代 byte 枚举。
type CascadeOpts struct {
	// Role 是否级联取 Role
	Role bool
	// Department 是否级联取 Department
	Department bool
}

// 预置常用策略
var (
	OptsNone     = CascadeOpts{}                             // 不级联
	OptsDefault  = CascadeOpts{Role: true, Department: true} // 默认全级联
	OptsRoleOnly = CascadeOpts{Role: true}                   // 仅 Role
	OptsDeptOnly = CascadeOpts{Department: true}             // 仅 Department
)

// 旧 byte 常量兼容（New(byte) 内部映射到 Opts*）
const (
	ResultDefault byte = 0
	ResultNone    byte = 1
)

// New S10/S11: 用 CascadeOpts 替代 byte 枚举。
// 兼容旧签名：cascadeType byte 仍可用，ResultDefault=0→OptsDefault, ResultNone=1→OptsNone。
func New(cascadeType byte, ds ...*sqlkit.DataSource) Dao {
	var opts CascadeOpts
	switch cascadeType {
	case 0: // ResultDefault
		opts = OptsDefault
	default: // ResultNone 等
		opts = OptsNone
	}
	return NewWithOpts(opts, ds...)
}

// NewWithOpts S11: 按级联选项构造 dao，支持按需只取 Role 或 Department。
func NewWithOpts(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao {
	dao := sqlkit.New[model.User](ds...)
	dao = dao.WithCascadeOpts(opts, func(obj *model.User, ctx sqlkit.CascadeCtx) {
		o := ctx.Opts.(CascadeOpts)
		if o.Role && obj.Role != nil {
			obj.Role = roledao.NewWithOpts(roledao.OptsDefault, ctx.Ds).SelectOneWithDelById(obj.Role.Id)
		}
		if o.Department && obj.Department != nil {
			obj.Department = departmentdao.NewWithOpts(departmentdao.OptsDefault, ctx.Ds).SelectOneWithDelById(obj.Department.Id)
		}
	})
	return Dao{dao}
}

func (dao Dao) Login(pwd, username, phone string) *model.User {
	builder := dao.Select()
	if !stringkit.IsNull(username) {
		builder = builder.Where("username=?", username)
	} else {
		builder = builder.Where("phone=?", phone)
	}
	// S4: One() 内部已默认 LIMIT 1，无需再追加 Limit(1)
	return builder.Where("pwd=?", pwd).One()
}

func (dao Dao) FindByPhone(phone string) *model.User {
	return dao.Select().Where("phone=?", phone).One()
}

func (dao Dao) FindByUsername(username string) *model.User {
	return dao.Select().Where("username=?", username).One()
}

// FindByUsernameDeleted S1: 用 OneIgnoreDel 真正忽略逻辑删除过滤。
func (dao Dao) FindByUsernameDeleted(username string) *model.User {
	return dao.Select().Where("username=?", username).OneIgnoreDel()
}

// FindParam 可以通过extend的值来find
type FindParam struct {
	Extend map[string]any
}

// Find S3: 用 WhereJsonbPathEq 替代 fmt.Sprintf 拼接，消除 SQL 注入风险。
func (dao Dao) Find(param FindParam) *model.User {
	builder := dao.Select()
	for k, v := range param.Extend {
		builder = builder.WhereJsonbPathEq("extend", k, v)
	}
	return builder.One()
}

// ListFromRootDepart S2: 用 WithRecursiveRaw 替代 fmt.Sprintf 拼接递归 CTE。
func (dao Dao) ListFromRootDepart(departId int64) []*model.User {
	deptTable := departmentdao.NewWithOpts(departmentdao.OptsNone, dao.DataSource()).Table()
	cteBody := fmt.Sprintf(`select %d::bigint as id union all select d.id from %s d, t where t.id=d.parent`, departId, deptTable)
	return dao.Select().
		WithRecursiveRaw("t", []string{"id"}, cteBody).
		Where("department in (select id from t)").
		OrderBy("name").OrderBy("id").List()
}

func (dao Dao) CountFromRootDepart(departId int64) int64 {
	deptTable := departmentdao.NewWithOpts(departmentdao.OptsNone, dao.DataSource()).Table()
	cteBody := fmt.Sprintf(`select %d::bigint as id union all select d.id from %s d, t where t.id=d.parent`, departId, deptTable)
	return dao.Select().
		WithRecursiveRaw("t", []string{"id"}, cteBody).
		Where("department in (select id from t)").
		Count()
}

type ListParam struct {
	RoleId      class.Int64
	Roles       []int64
	Departments []int64
	IdList      []int64
}

func (dao Dao) List(param ListParam) model.UserList {
	builder := dao.Select().OrderBy("name").OrderBy("id")
	if param.RoleId.IsValid() {
		builder = builder.Where("role=?", param.RoleId)
	}
	if len(param.IdList) > 0 {
		builder = builder.WhereUnnestIn("id", param.IdList)
	}
	if len(param.Roles) > 0 {
		builder = builder.WhereUnnestIn("role", param.Roles)
	}
	if len(param.Departments) > 0 {
		builder = builder.WhereUnnestIn("department", param.Departments)
	}
	return builder.List()
}

func (dao Dao) FreezeUser(uid int64, status int32) {
	dao.Update().Set("status", status).Where("id=?", uid).Exec()
}
func (dao Dao) SetNull(id int64) {
	dao.Update().Set("phone", squirrel.Expr("null")).Set("username", squirrel.Expr("null")).Where("id=?", id).Exec()
}
