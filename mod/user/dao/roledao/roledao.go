package roledao

import (
	"fmt"

	"github.com/example/go-ai-scaffold/mod/user/dao/departmentdao"
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

type Dao struct {
	sqlkit.Dao[model.Role]
}

// CascadeOpts S10/S11: 级联策略选项，替代 byte 枚举。
type CascadeOpts struct {
	// Department 是否级联取 Department
	Department bool
}

var (
	OptsNone    = CascadeOpts{}                 // 不级联
	OptsDefault = CascadeOpts{Department: true} // 默认级联 Department
)

// New 按 CascadeOpts 构造 dao。
func New(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao {
	dao := sqlkit.New[model.Role](ds...)
	dao = dao.WithCascadeOpts(opts, func(obj *model.Role, ctx sqlkit.CascadeCtx) {
		o := ctx.Opts.(CascadeOpts)
		if o.Department && obj.Department != nil {
			obj.Department = departmentdao.New(departmentdao.OptsDefault, ctx.Ds).SelectOneWithDelById(obj.Department.Id)
		}
	})
	return Dao{dao}
}

func (dao Dao) FindByName(name string) *model.Role {
	// S4: One() 内部已默认 LIMIT 1
	return dao.Select().Where("name=?", name).One()
}

// ListFromRootDepart S2: 用 WithRecursiveRaw 替代 fmt.Sprintf 拼接递归 CTE。
func (dao Dao) ListFromRootDepart(id int64) []*model.Role {
	deptTable := departmentdao.New(departmentdao.OptsNone, dao.DataSource()).Table()
	cteBody := fmt.Sprintf(`select %d::bigint as id union all select d.id from %s d, t where t.id=d.parent`, id, deptTable)
	return dao.Select().
		WithRecursiveRaw("t", []string{"id"}, cteBody).
		Where("id>0 and department in (select id from t)").
		OrderBy("id").List()
}

func (dao Dao) CountFromRootDepart(id int64) int64 {
	deptTable := departmentdao.New(departmentdao.OptsNone, dao.DataSource()).Table()
	cteBody := fmt.Sprintf(`select %d::bigint as id union all select d.id from %s d, t where t.id=d.parent`, id, deptTable)
	return dao.Select().
		WithRecursiveRaw("t", []string{"id"}, cteBody).
		Where("department in (select id from t)").
		Count()
}

type ListParam struct {
	Departments []int64
}

func (dao Dao) List(param ListParam) []*model.Role {
	builder := dao.Select().Where("id>0 and department>=0").OrderBy("id")
	if len(param.Departments) > 0 {
		builder = builder.WhereUnnestIn("department", param.Departments)
	}
	return builder.List()
}

func (dao Dao) Count(param ListParam) int64 {
	builder := dao.Select()
	if len(param.Departments) > 0 {
		builder = builder.WhereUnnestIn("department", param.Departments)
	}
	return builder.Count()
}
