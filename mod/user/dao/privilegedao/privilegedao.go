package privilegedao

import (
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

type Dao struct {
	sqlkit.Dao[model.PrivilegeConstant]
}

// CascadeOpts 统一级联策略签名。PrivilegeConstant 无级联关系，opts 仅用于签名一致性。
type CascadeOpts struct{}

var OptsNone = CascadeOpts{}

// New 按 CascadeOpts 构造 dao。与其它 dao 统一签名，opts 当前不使用。
func New(opts CascadeOpts, ds ...*sqlkit.DataSource) Dao {
	return Dao{sqlkit.New[model.PrivilegeConstant](ds...)}
}

func (dao Dao) ListPrivileges() []*model.PrivilegeConstant {
	return dao.Select().OrderBy("sort").List()
}
