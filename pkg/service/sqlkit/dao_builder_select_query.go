package sqlkit

import (
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cast"
)

// 链式查询的终结方法

func (dao SelectDao[T]) QueryRows() *sqlx.Rows {
	sql, args := dao.Sql()
	return dao.QueryRaw(sql, args)
}

// One 取一条。S4: 内部默认 LIMIT 1，调用方无需再显式追加。
// S6: 迭代后检查 rows.Err()，区分"无数据"与"查询失败"。
func (dao SelectDao[T]) One() *T {
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	d = d.Limit(1)
	rows := d.QueryRows()
	defer rows.Close()
	for rows.Next() {
		m := scanStruct[T](rows, dao.dataSource.Driver)
		if d.Cascade != nil {
			d.Cascade(m)
		}
		return m
	}
	if err := rows.Err(); err != nil {
		panic(exception.New(err.Error()))
	}
	return nil
}

func (dao SelectDao[T]) List() []*T {
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	return scanObjList(d)
}

func (dao SelectDao[T]) OneMap() map[string]any {
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	d = d.Limit(1)
	rows := d.QueryRows()
	defer rows.Close()
	for rows.Next() {
		m := map[string]any{}
		err := rows.MapScan(m)
		if err != nil {
			panic(exception.New(err.Error()))
		}
		return m
	}
	if err := rows.Err(); err != nil {
		panic(exception.New(err.Error()))
	}
	return nil
}

// OneString 取一个string值
func (dao SelectDao[T]) OneString() string {
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	d = d.Limit(1)
	rows := d.QueryRows()
	defer rows.Close()
	for rows.Next() {
		ret, err := rows.SliceScan()
		if err != nil {
			panic(exception.New(err.Error()))
		}
		return cast.ToString(ret[0])
	}
	if err := rows.Err(); err != nil {
		panic(exception.New(err.Error()))
	}
	return ""
}

// OneNumber 取一个number值
func (dao SelectDao[T]) OneNumber() int64 {
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	d = d.Limit(1)
	rows := d.QueryRows()
	defer rows.Close()
	for rows.Next() {
		ret, err := rows.SliceScan()
		if err != nil {
			panic(exception.New(err.Error()))
		}
		return cast.ToInt64(ret[0])
	}
	if err := rows.Err(); err != nil {
		panic(exception.New(err.Error()))
	}
	return 0
}

// Count 计数值
func (dao SelectDao[T]) Count() int64 {
	d := dao.resetColumns("count(1)")
	return d.OneNumber()
}

func (dao SelectDao[T]) ListMap() []map[string]any {
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	rows := d.QueryRows()
	defer rows.Close()
	list := make([]map[string]any, 0, 5)
	for rows.Next() {
		m := map[string]any{}
		err := rows.MapScan(m)
		if err != nil {
			panic(exception.New(err.Error()))
		}
		list = append(list, m)
	}
	if err := rows.Err(); err != nil {
		panic(exception.New(err.Error()))
	}
	return list
}

// S5: 移除重复的 defer rows.Close()
func (dao SelectDao[T]) ListString() []string {
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	rows := d.QueryRows()
	defer rows.Close()
	list := make([]string, 0, 5)
	for rows.Next() {
		ret, err := rows.SliceScan()
		if err != nil {
			panic(exception.New(err.Error()))
		}
		list = append(list, cast.ToString(ret[0]))
	}
	if err := rows.Err(); err != nil {
		panic(exception.New(err.Error()))
	}
	return list
}

type Page struct {
	PageSize uint64 // 一页数量
	PageNum  uint64 // 第几页
}

// Page 分页：返回数据和总数量
// S7: count 路径用 resetColumns("1") 减少子查询数据量；原查询的 where 条件保留。
func (dao SelectDao[T]) Page(p Page) ([]*T, uint64) {
	if !(p.PageSize > 0 && p.PageNum > 0) {
		panic(exception.New("page 参数范围错误"))
	}
	d := dao
	if !dao.ignoreLogicDel {
		d = dao.whereNLogicDel()
	}
	// 分页数据
	d1 := d.Limit(p.PageSize).Offset(p.PageSize * (p.PageNum - 1))
	// 总数：内层只取常量列 1，避免回传全部字段
	d2 := d.resetColumns("1").Prefix("select count(1) from (").Suffix(") t")
	return scanObjList(d1), cast.ToUint64(d2.OneString())
}
