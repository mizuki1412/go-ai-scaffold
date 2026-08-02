package departmentdao

import (
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

type Dao struct {
	sqlkit.Dao[model.Department]
}

// CascadeOpts S10/S11: 级联策略选项，替代 byte 枚举。
type CascadeOpts struct {
	// Children 是否级联取子部门
	Children bool
	// Parent 是否级联取父部门
	Parent bool
}

var (
	OptsNone     = CascadeOpts{}                             // 不级联
	OptsDefault  = CascadeOpts{Parent: true}                 // 默认级联父部门
	OptsChildren = CascadeOpts{Children: true}               // 仅子部门
	OptsAll      = CascadeOpts{Children: true, Parent: true} // 子+父
)

// 旧 byte 常量兼容（New(byte) 内部映射到 Opts*）
const (
	ResultDefault  byte = 0
	ResultChildren byte = 1
	ResultAll      byte = 2
	ResultNone     byte = 3
)

// New 兼容旧签名：cascadeType byte 仍可用。
// ResultDefault=0→OptsDefault, ResultChildren=1→OptsChildren, ResultAll=2→OptsAll, ResultNone=3→OptsNone
func New(cascadeType byte, ds ...*sqlkit.DataSource) Dao {
	var opts CascadeOpts
	switch cascadeType {
	case 0: // ResultDefault
		opts = OptsDefault
	case 1: // ResultChildren
		opts = OptsChildren
	case 2: // ResultAll
		opts = OptsAll
	default: // ResultNone
		opts = OptsNone
	}
	return NewWithOpts(opts, ds...)
}

// NewWithOpts S11: 按级联选项构造 dao。
func NewWithOpts(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao {
	d := sqlkit.New[model.Department](ds...)
	dao := Dao{d}
	dao.Dao = dao.WithCascadeOpts(opts, func(obj *model.Department, ctx sqlkit.CascadeCtx) {
		o := ctx.Opts.(CascadeOpts)
		// 级联取子部门/父部门时，用 OptsNone 的 dao 避免无限递归
		if o.Children {
			obj.Children = NewWithOpts(OptsNone, ctx.Ds).ListByParent(obj.Id)
		}
		if o.Parent && obj.Parent != nil {
			obj.Parent = NewWithOpts(OptsDefault, ctx.Ds).SelectOneWithDelById(obj.Parent.Id)
		}
		if !o.Children {
			obj.Children = nil
		}
		if !o.Parent {
			obj.Parent = nil
		}
	})
	return dao
}

func (dao Dao) ListByParent(id int64) model.DeptList {
	return dao.Select().Where("parent=?", id).OrderBy("no").OrderBy("id").List()
}

func (dao Dao) ListAll() model.DeptList {
	return dao.Select().Where("id>0").OrderBy("parent").OrderBy("no").OrderBy("id").List()
}

func (dao Dao) FindByName(name string) *model.Department {
	// S4: One() 内部已默认 LIMIT 1
	return dao.Select().Where("name=?", name).One()
}

func (dao Dao) FindByNo(no string) *model.Department {
	// S4: One() 内部已默认 LIMIT 1
	return dao.Select().Where("no=?", no).One()
}
