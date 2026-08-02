package privilegedao

import (
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/pkg/service/sqlkit"
)

type Dao struct {
	sqlkit.Dao[model.PrivilegeConstant]
}

func New(ds ...*sqlkit.DataSource) Dao {
	return Dao{sqlkit.New[model.PrivilegeConstant](ds...)}
}

func (dao Dao) ListPrivileges() []*model.PrivilegeConstant {
	return dao.Select().OrderBy("sort").List()
}
