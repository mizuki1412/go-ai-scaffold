package sqlkit

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/const/sqlconst"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/spf13/cast"
)

func (dao Dao[T]) InsertObj(dest *T) {
	var columns []string
	var vals []any
	rv := reflect.ValueOf(dest).Elem()
	for _, e := range dao.modelMeta.allInsertKeys {
		var val = e.val(rv, dao.dataSource.Driver)
		if val == nil {
			continue
		}
		if sqlconst.IsTaos(dao.dataSource.Driver) {
			columns = append(columns, dao.dataSource.EscapeName(e.OriKey))
		} else {
			columns = append(columns, e.OriKey)
		}
		vals = append(vals, val)
	}
	if len(columns) == 0 {
		panic(exception.New("no fields", 2))
	}
	if sqlconst.IsTaos(dao.dataSource.Driver) {
		// 针对taos重写
		vals = argsWrap(dao.dataSource.Driver, vals)
		valPlaceholders := make([]string, 0, len(vals))
		for _, e := range vals {
			switch e.(type) {
			case string, class.String:
				valPlaceholders = append(valPlaceholders, "'?'")
			default:
				valPlaceholders = append(valPlaceholders, "?")
			}
		}
		ss := fmt.Sprintf("insert into %s(%s) values(%s)",
			dao.modelMeta.getTable(), strings.Join(columns, ", "), strings.Join(valPlaceholders, ", "))
		res := dao.ExecRaw(ss, vals)
		rn, _ := res.RowsAffected()
		logkit.Debug("sql res", "rows", rn)
	} else if sqlconst.IsPostgresType(dao.dataSource.Driver) || dao.dataSource.Driver == sqlconst.Sqlite3 {
		// PG/Kingbase/SQLite 支持 RETURNING *，可一次拿到插入后的完整行（含自增主键）
		builder := dao.Insert()
		builder = builder.Columns(columns...).Values(vals...)
		builder = builder.Suffix("returning *")
		builder.ReturnOne(dest)
	} else {
		// MySQL/SQL Server/Oracle/DM 等不支持 RETURNING *，仅执行插入；
		// 自增主键不会回填到 dest，调用方需要另行获取（如 LastInsertId / 二次查询）。
		builder := dao.Insert()
		builder = builder.Columns(columns...).Values(vals...)
		builder.Exec()
	}
}

func (dao Dao[T]) InsertBatch(dest []*T) {
	if len(dest) == 0 {
		panic(exception.New("insert batch need dest"))
	}
	// 先收集所有行字段的并集，保证每行 vals 与 columns 严格对齐。
	// 否则当某行字段为 nil 被跳过、而其他行该字段有值时，列与值会错位。
	rvs := make([]reflect.Value, 0, len(dest))
	colKeys := make([]ModelMetaKey, 0)
	colIndex := make(map[string]int)
	for _, e := range dest {
		rv := reflect.ValueOf(e).Elem()
		rvs = append(rvs, rv)
		for _, k := range dao.modelMeta.allInsertKeys {
			if k.val(rv, dao.dataSource.Driver) == nil {
				continue
			}
			if _, ok := colIndex[k.OriKey]; !ok {
				colIndex[k.OriKey] = len(colKeys)
				colKeys = append(colKeys, k)
			}
		}
	}
	if len(colKeys) == 0 {
		panic(exception.New("no fields", 2))
	}
	// 按并集对齐每行 vals，缺失字段补 nil
	valsArr := make([][]any, 0, len(rvs))
	for _, rv := range rvs {
		row := make([]any, len(colKeys))
		for _, k := range dao.modelMeta.allInsertKeys {
			val := k.val(rv, dao.dataSource.Driver)
			if val == nil {
				continue
			}
			row[colIndex[k.OriKey]] = val
		}
		valsArr = append(valsArr, row)
	}
	if sqlconst.IsTaos(dao.dataSource.Driver) {
		var columns []string
		for _, k := range colKeys {
			columns = append(columns, dao.dataSource.EscapeName(k.OriKey))
		}
		sql := ""
		var allVals []any
		for i := 0; i < len(valsArr); i++ {
			vals := argsWrap(dao.dataSource.Driver, valsArr[i])
			valPlaceholders := make([]string, 0, len(vals))
			for _, e := range vals {
				switch e.(type) {
				case string, class.String:
					valPlaceholders = append(valPlaceholders, "'?'")
				default:
					valPlaceholders = append(valPlaceholders, "?")
				}
			}
			if i == 0 {
				sql += fmt.Sprintf("insert into %s(%s) values", dao.modelMeta.getTable(), strings.Join(columns, ", "))
			}
			sql += fmt.Sprintf("(%s)", strings.Join(valPlaceholders, ", "))
			if i < len(valsArr)-1 {
				sql += ", "
			}
			allVals = append(allVals, vals...)
		}
		res := dao.ExecRaw(sql, allVals)
		rn, _ := res.RowsAffected()
		logkit.Debug("sql res", "rows", rn)
	} else {
		var columns []string
		for _, k := range colKeys {
			columns = append(columns, k.OriKey)
		}
		builder := dao.Insert()
		builder = builder.Columns(columns...)
		for i := range valsArr {
			builder = builder.Values(valsArr[i]...)
		}
		builder.Exec()
	}
}

func (dao Dao[T]) UpdateObj(dest *T) int64 {
	builder := dao.Update()
	rv := reflect.ValueOf(dest).Elem()
	for _, e := range dao.modelMeta.allUpdateKeys {
		var val = e.val(rv, dao.dataSource.Driver)
		if val == nil {
			continue
		}
		// 针对class.MapString 采用merge方式 todo mysql
		if (e.RStruct.Type.String() == "class.MapString" || e.RStruct.Type.String() == "class.MapStringSync") && sqlconst.IsPostgresType(dao.dataSource.Driver) {
			builder = builder.Set(e.OriKey, squirrel.Expr("coalesce("+e.OriKey+",'{}'::jsonb) || ?", val))
		} else {
			builder = builder.Set(e.OriKey, val)
		}
	}
	for _, e := range dao.modelMeta.allPKs {
		v := e.val(rv, dao.dataSource.Driver)
		if v == nil {
			panic(exception.New("pk val is nil"))
		}
		builder = builder.Where(e.Key+"=?", v)
	}
	return builder.Exec()
}

func (dao Dao[T]) DeleteById(id ...any) int64 {
	if len(id) != len(dao.modelMeta.allPKs) {
		panic(exception.New("主键数量不匹配"))
	}
	if dao.modelMeta.logicDelKey.Key != "" {
		builder := dao.Update()
		builder = builder.Set(dao.modelMeta.logicDelKey.OriKey, builder.LogicDelVal[0])
		for i := 0; i < len(dao.modelMeta.allPKs); i++ {
			builder = builder.Where(dao.modelMeta.allPKs[i].Key+"=?", id[i])
		}
		return builder.Exec()
	} else {
		builder := dao.Delete()
		for i := 0; i < len(dao.modelMeta.allPKs); i++ {
			builder = builder.Where(dao.modelMeta.allPKs[i].Key+"=?", id[i])
		}
		return builder.Exec()
	}
}

// SelectOneById 根据id获取，计算逻辑删除
func (dao Dao[T]) SelectOneById(id ...any) *T {
	builder := dao.Select()
	if len(id) != len(dao.modelMeta.allPKs) {
		panic(exception.New("主键数量不匹配"))
	}
	for i := 0; i < len(dao.modelMeta.allPKs); i++ {
		builder = builder.Where(dao.modelMeta.allPKs[i].Key+"=?", id[i])
	}
	return builder.One()
}

// SelectOneWithDelById 根据id获取，忽略逻辑删除
func (dao Dao[T]) SelectOneWithDelById(id ...any) *T {
	builder := dao.Select()
	if len(id) != len(dao.modelMeta.allPKs) {
		panic(exception.New("主键数量不匹配"))
	}
	for i := 0; i < len(dao.modelMeta.allPKs); i++ {
		builder = builder.Where(dao.modelMeta.allPKs[i].Key+"=?", id[i])
	}
	return builder.IgnoreLogicDel().One()
}

// CheckSchemaExist schema是否存在
func (dao Dao[T]) CheckSchemaExist(schema string) bool {
	if dao.dataSource.Driver == sqlconst.Postgres {
		// 参数化查询，避免 schema 名含特殊字符引发注入
		p1 := rawPlaceholder(dao.dataSource.Driver, 1)
		rows := dao.QueryRaw(
			"SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = "+p1+")",
			[]any{schema})
		defer rows.Close()
		for rows.Next() {
			ret, err := rows.SliceScan()
			if err != nil {
				panic(exception.New(err.Error()))
			}
			return len(ret) > 0 && cast.ToBool(ret[0])
		}
	}
	return false
}

// CheckTableExist 检查表是否存在
func (dao Dao[T]) CheckTableExist(t string) bool {
	if dao.dataSource.Driver == sqlconst.Sqlite3 {
		rows := dao.QueryRaw(
			"SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?",
			[]any{t})
		defer rows.Close()
		for rows.Next() {
			ret, err := rows.SliceScan()
			if err != nil {
				panic(exception.New(err.Error()))
			}
			return len(ret) > 0 && cast.ToInt32(ret[0]) >= 1
		}
	} else {
		// todo other db
		panic(exception.New("CheckTableExist在此数据库未实现"))
	}
	return false
}
