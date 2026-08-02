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

// New 按 CascadeOpts 构造 dao。
// S13: 用 WithCascadeBatchLinks + 声明式 link 替代手写收集/分发逻辑。
// 每个 link 描述一个 T→*U 级联关系，sqlkit 统一处理"收集 id → 批量 IN 查询 → 按 id 分发"。
// Role/Department 内部用各自的 OptsDefault，会递归走批量级联（role 再批量取 department）。
func New(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao {
	dao := sqlkit.New[model.User](ds...)
	var links []sqlkit.CascadeLinker[model.User]
	if opts.Role {
		links = append(links, sqlkit.NewCascadeLink(
			func(u *model.User) *model.Role { return u.Role },
			func(u *model.User, r *model.Role) { u.Role = r },
			func(r *model.Role) int64 { return r.Id },
			func(ids []int64, ds *sqlkit.DataSource) []*model.Role {
				return roledao.New(roledao.OptsDefault, ds).SelectByIdsIgnoreDel(ids)
			},
		))
	}
	if opts.Department {
		links = append(links, sqlkit.NewCascadeLink(
			func(u *model.User) *model.Department { return u.Department },
			func(u *model.User, d *model.Department) { u.Department = d },
			func(d *model.Department) int64 { return d.Id },
			func(ids []int64, ds *sqlkit.DataSource) []*model.Department {
				return departmentdao.New(departmentdao.OptsDefault, ds).SelectByIdsIgnoreDel(ids)
			},
		))
	}
	return Dao{dao.WithCascadeBatchLinks(opts, links...)}
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
// B22: 参数化 rootId，去掉 PG 专用的 ::bigint cast，兼容多 driver。
func (dao Dao) ListFromRootDepart(departId int64) []*model.User {
	deptTable := departmentdao.New(departmentdao.OptsNone, dao.DataSource()).Table()
	cteBody := fmt.Sprintf(`select ? as id union all select d.id from %s d, t where t.id=d.parent`, deptTable)
	return dao.Select().
		WithRecursiveRaw("t", []string{"id"}, cteBody, departId).
		Where("department in (select id from t)").
		OrderBy("name").OrderBy("id").List()
}

func (dao Dao) CountFromRootDepart(departId int64) int64 {
	deptTable := departmentdao.New(departmentdao.OptsNone, dao.DataSource()).Table()
	cteBody := fmt.Sprintf(`select ? as id union all select d.id from %s d, t where t.id=d.parent`, deptTable)
	return dao.Select().
		WithRecursiveRaw("t", []string{"id"}, cteBody, departId).
		Where("department in (select id from t)").
		Count()
}

type ListParam struct {
	RoleId      class.Int64
	Roles       []int64
	Departments []int64
	IdList      []int64
	// B21: 分页参数，PageSize=0 时不分页
	Page *sqlkit.Page
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
	if param.Page != nil && param.Page.PageSize > 0 {
		list, _ := builder.Page(*param.Page)
		return list
	}
	return builder.List()
}

func (dao Dao) FreezeUser(uid int64, status int32) {
	dao.Update().Set("status", status).Where("id=?", uid).Exec()
}
func (dao Dao) SetNull(id int64) {
	dao.Update().Set("phone", squirrel.Expr("null")).Set("username", squirrel.Expr("null")).Where("id=?", id).Exec()
}
