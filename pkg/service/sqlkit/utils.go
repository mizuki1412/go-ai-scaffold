package sqlkit

import (
	"reflect"

	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/constraints"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/cli/tag"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cast"
)

func scanObjList[T any](dao SelectDao[T]) []*T {
	rows := dao.QueryRows()
	list := make([]*T, 0, 5)
	defer rows.Close()
	for rows.Next() {
		list = append(list, scanStruct[T](rows, dao.dataSource.Driver))
	}
	// S6: 检查迭代错误，区分"无数据"与"查询失败"
	if err := rows.Err(); err != nil {
		panic(exception.New(err.Error()))
	}
	if dao.Cascade != nil {
		for i := range list {
			dao.Cascade(list[i])
		}
	}
	return list
}

func scanStruct[T any](rows *sqlx.Rows, driver string) *T {
	m := new(T)
	err := rows.StructScan(m)
	rv := reflect.ValueOf(m).Elem()
	rt := reflect.TypeOf(m).Elem()
	// S18: scanStruct 作为泛型函数无法持有 dao 上下文，仍需全字段 reflect。
	// ModelMeta.scanFields 预计算已就绪，待后续 scan 路径接入 dao 实例后启用。
	for i := 0; i < rv.NumField(); i++ {
		v := rv.Field(i)
		if v.Kind() == reflect.Struct {
			obj := v.Addr().Interface()
			// 处理 arr, 只针对 struct; 设置 dbdriver
			if vv, ok := obj.(constraints.SetDBDriverInterface); ok {
				vv.SetDBDriver(driver)
			}
			// 对decimal精度的处理
			precision := cast.ToInt32(rt.Field(i).Tag.Get(tag.DecimalPrecision.Name))
			if precision > 0 {
				if vv, ok := obj.(class.Decimal); ok {
					vv.Set(vv.Round(precision))
				}
				if vv, ok := obj.(*class.Decimal); ok {
					vv.Set(vv.Round(precision))
				}
			}
		}
	}
	if err != nil {
		panic(exception.New(err.Error()))
	}
	return m
}
