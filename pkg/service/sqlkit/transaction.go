package sqlkit

import (
	"github.com/example/go-ai-scaffold/pkg/library/c"
)

// TxArea 事务物理代码块，不指定datasource时，用defaultDataSource
// S17: 保留原始异常类型与调用栈，不再重新包装成 exception.New(ex.Msg) 丢失类型。
func TxArea(f func(targetDS *DataSource), dataSources ...*DataSource) {
	var ds *DataSource
	if len(dataSources) == 0 {
		ds = DefaultDataSource()
	} else {
		ds = dataSources[0]
	}
	ex := c.RecoverFuncWrapper(func() {
		ds.BeginTX()
		// 传入带tx的datasource，内部代码用这个ds
		f(ds)
		ds.Commit()
	})
	if ex != nil {
		ds.Rollback()
		panic(ex) // 保留原始异常
	}
}

// TxAreaMulti 多数据源事务块。S16: 支持多数据源各自开启事务，
// 任一执行失败则回滚所有已开启未提交的数据源；全部成功才逐个提交。
// f 接收按入参顺序、已开启事务的数据源切片。
func TxAreaMulti(dataSources []*DataSource, f func(dss []*DataSource)) {
	if len(dataSources) == 0 {
		dataSources = []*DataSource{DefaultDataSource()}
	}
	opened := make([]*DataSource, 0, len(dataSources))
	committed := 0
	ex := c.RecoverFuncWrapper(func() {
		for _, ds := range dataSources {
			ds.BeginTX()
			opened = append(opened, ds)
		}
		f(opened)
		for _, ds := range opened {
			ds.Commit()
			committed++
		}
	})
	if ex != nil {
		// 回滚 committed 之后（即未成功提交）的数据源
		for i := committed; i < len(opened); i++ {
			opened[i].Rollback()
		}
		panic(ex)
	}
}
