package class

import (
	"github.com/example/go-ai-scaffold/pkg/library/timekit"
)

// Any 系列方法：nullable 包装类型取底层值供 map[string]any 组装（响应构建等场景）。
// 约定：无效（DB NULL / 未赋值）时返回 nil（序列化为 JSON null）；有效时返回底层零类型值
// （string/int64/int32/float64/bool/时间串/decimal）。
// 业务代码构建响应 map 时不得直接取 .String/.Int64/.Int32 等裸值——NULL 会退化为
// ""/0，与 PHP 原接口（FastAdmin）返回 null 的字段语义不一致；统一改用 .Any()。

// Any 返回底层字符串；无效时返回 nil。
func (th String) Any() any {
	if th.Valid {
		return th.String
	}
	return nil
}

// Any 返回底层 int64；无效时返回 nil。
func (th Int64) Any() any {
	if th.Valid {
		return th.Int64
	}
	return nil
}

// Any 返回底层 int32；无效时返回 nil。
func (th Int32) Any() any {
	if th.Valid {
		return th.Int32
	}
	return nil
}

// Any 返回底层 float64；无效时返回 nil。
func (th Float64) Any() any {
	if th.Valid {
		return th.Float64
	}
	return nil
}

// Any 返回底层 bool；无效时返回 nil。
func (th Bool) Any() any {
	if th.Valid {
		return th.Bool
	}
	return nil
}

// Any 返回格式化时间串（与 MarshalJSON 同布局）；无效时返回 nil。
func (th Time) Any() any {
	if th.Valid {
		return th.Time.Format(timekit.TimeLayout2)
	}
	return nil
}

// Any 返回底层 decimal；无效时返回 nil（序列化交由 decimal 自身 MarshalJSON）。
func (th Decimal) Any() any {
	if th.Valid {
		return th.Decimal
	}
	return nil
}

// Any 返回底层切片（无效时切片本身为 nil，序列化为 null）。
func (th ArrInt) Any() any {
	return th.Array
}

// Any 返回底层 map（无效时 map 本身为 nil，序列化为 null）。
func (th MapString) Any() any {
	return th.Map
}
