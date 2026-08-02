package sqlkit

import (
	"database/sql"

	"github.com/example/go-ai-scaffold/pkg/class/const/sqlconst"
	"github.com/example/go-ai-scaffold/pkg/library/jsonkit"
	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/jmoiron/sqlx"
)

// LogicDelVal 全局逻辑删除的 value
var LogicDelVal = []any{true, false}

type Dao[T any] struct {
	meta T
	// 逻辑删除的字段，可替代全局的LogicDelVal
	LogicDelVal []any
	// 级联实现的函数
	Cascade func(*T)
	// 数据源
	dataSource *DataSource
	// 目标表结构
	modelMeta ModelMeta
}

type DaoModelMeta interface {
	getModelMeta() ModelMeta
	// S20: Join 等方法需要拿到目标 dao 的数据源以构造带 schema 的表名
	getDataSource() *DataSource
}

// New 必须从初始化函数生成 dao
func New[T any](ds ...*DataSource) Dao[T] {
	dao := Dao[T]{}
	if len(ds) > 0 {
		dao.dataSource = ds[0]
	} else {
		dao.dataSource = DefaultDataSource()
	}
	dao.modelMeta = dao.modelMeta.init(dao.meta, dao.dataSource)
	return dao
}

func (dao Dao[T]) getModelMeta() ModelMeta {
	return dao.modelMeta
}

// S20: 暴露数据源，供 Join 等方法构造目标表名
func (dao Dao[T]) getDataSource() *DataSource {
	return dao.dataSource
}

func (dao Dao[T]) DataSource() *DataSource {
	return dao.dataSource
}

func (dao Dao[T]) QueryRaw(sql string, args []any) *sqlx.Rows {
	logkit.Debug("sql req", "sql", sql, "args", jsonkit.ToString(args))
	return dao.dataSource.Query(sql, args)
}

// QueryRawRows 默认返回 T list，可用于自由sql时，自定义返回值
func (dao Dao[T]) QueryRawRows(sql string, args []any) []*T {
	rows := dao.QueryRaw(sql, args)
	list := make([]*T, 0, 5)
	defer rows.Close()
	for rows.Next() {
		list = append(list, scanStruct[T](rows, dao.dataSource.Driver))
	}
	if dao.Cascade != nil {
		for i := range list {
			dao.Cascade(list[i])
		}
	}
	return list
}

func (dao Dao[T]) ExecRaw(sql string, args []any) sql.Result {
	logkit.Debug("sql exec", "sql", sql, "args", jsonkit.ToString(args))
	return dao.dataSource.Exec(sql, args)
}

func (dao Dao[T]) GetOriginDB() *sqlx.DB {
	return dao.dataSource.DBPool
}

// Close 程序退出时的逻辑
func (dao Dao[T]) Close() {
	if dao.dataSource.Driver == sqlconst.Sqlite3 {
		dao.GetOriginDB().Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	}
	dao.GetOriginDB().Close()
}

/// 小功能

func (dao Dao[T]) Table(alias ...string) string {
	return dao.modelMeta.getTable(dao.dataSource, alias...)
}
func (dao Dao[T]) EscapeNames(name ...string) []string {
	return dao.modelMeta.escapeNames(dao.dataSource, name)
}

// WithSchema S9: 返回一个使用指定 schema 的新 Dao，复用原 ModelMeta 缓存与 DBPool。
// 用于请求级 schema 注入，替代业务层反复 `dao.DataSource().Schema = ...`。
func (dao Dao[T]) WithSchema(schema string) Dao[T] {
	dao.dataSource = dao.dataSource.WithSchema(schema)
	return dao
}

// ============ S10/S11: Cascade 级联策略抽象 ============
// CascadeCtx 提供级联执行的上下文，供 CascadeFunc 决定是否执行某级联。
// S11: 替代单一 func(*T)，支持按需控制级联粒度。
type CascadeCtx struct {
	// Opts 为调用方传入的级联选项（由业务自定义）
	Opts any
	// Ds 为当前数据源，级联查询可用
	Ds *DataSource
}

// CascadeFunc 是带上下文的级联函数签名。
type CascadeFunc[T any] func(obj *T, ctx CascadeCtx)

// WithCascade S10: 替换 dao 的级联函数，支持 CascadeCtx 传参。
// 与原 Cascade 字段并存，当 WithCascade 设置后，List/One 执行级联时会优先使用它。
func (dao Dao[T]) WithCascade(f CascadeFunc[T]) Dao[T] {
	dao.Cascade = func(obj *T) {
		f(obj, CascadeCtx{Ds: dao.dataSource})
	}
	return dao
}

// WithCascadeOpts S11: 带业务 opts 的级联设置。
// opts 会随 CascadeCtx.Opts 传入级联函数，调用方可在 opts 中编码"只取 Role 不取 Department"等策略。
func (dao Dao[T]) WithCascadeOpts(opts any, f CascadeFunc[T]) Dao[T] {
	dao.Cascade = func(obj *T) {
		f(obj, CascadeCtx{Opts: opts, Ds: dao.dataSource})
	}
	return dao
}
