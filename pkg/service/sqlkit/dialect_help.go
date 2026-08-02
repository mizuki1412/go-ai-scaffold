package sqlkit

import (
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/const/sqlconst"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
)

// 生成pg的array表达式
func pgArray(arr any) (string, []any) {
	var suffix string
	var args []any
	var flags []string
	switch arr.(type) {
	case []int:
		arr := arr.([]int)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int8:
		suffix = "int[]"
		arr := arr.([]int8)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int16:
		suffix = "int[]"
		arr := arr.([]int16)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int32:
		suffix = "int[]"
		arr := arr.([]int32)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int64:
		suffix = "bigint[]"
		arr := arr.([]int64)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []float32:
		suffix = "decimal[]"
		arr := arr.([]float32)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []float64:
		suffix = "decimal[]"
		arr := arr.([]float64)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []string:
		suffix = "varchar[]"
		arr := arr.([]string)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	default:
		panic(exception.New("pgArray params not supported"))
	}
	// 用{} 有错误：invalid input syntax for type integer
	return "ARRAY[" + strings.Join(flags, ",") + "]::" + suffix, args
}

func normalArray(arr any) (string, []any) {
	var args []any
	var flags []string
	switch arr.(type) {
	case []int:
		arr := arr.([]int)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int8:
		arr := arr.([]int8)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int16:
		arr := arr.([]int16)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int32:
		arr := arr.([]int32)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []int64:
		arr := arr.([]int64)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []float32:
		arr := arr.([]float32)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []float64:
		arr := arr.([]float64)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	case []string:
		arr := arr.([]string)
		flags = make([]string, len(arr))
		args = make([]any, len(arr))
		for i := 0; i < len(flags); i++ {
			flags[i] = "?"
			args[i] = arr[i]
		}
	default:
		panic(exception.New("normalArray params not supported"))
	}
	return "(" + strings.Join(flags, ",") + ")", args
}

// 占位符
func placeholder(driver string) squirrel.PlaceholderFormat {
	switch driver {
	case sqlconst.Postgres, sqlconst.Kingbase:
		return squirrel.Dollar
	case sqlconst.Oracle:
		return squirrel.Colon
	default:
		return squirrel.Question
	}
}

// rawPlaceholder 返回原生 SQL（不经过 squirrel）的占位符，避免字符串拼接引发的 SQL 注入。
// index 从 1 开始；对不区分位置的占位符（?）忽略 index。
func rawPlaceholder(driver string, index int) string {
	switch driver {
	case sqlconst.Postgres, sqlconst.Kingbase:
		return fmt.Sprintf("$%d", index)
	case sqlconst.Oracle, sqlconst.DM:
		return fmt.Sprintf(":%d", index)
	default:
		return "?"
	}
}

// args中部分值转换
// S19: 增加快速路径——若 args 不含任何需转换的类型，直接返回原切片，避免高频分配。
func argsWrap(driver string, args []any) []any {
	// todo 其他值类型
	sqliteOrTaos := driver == sqlconst.Sqlite3 || sqlconst.IsTaos(driver)
	if !sqliteOrTaos {
		// 非 sqlite/taos：仅需处理 class.Time / time.Time；先快速探测是否存在
		need := false
		for _, e := range args {
			switch e.(type) {
			case class.Time, time.Time:
				need = true
			}
			if need {
				break
			}
		}
		if !need {
			return args
		}
	}
	new_args := make([]any, 0, len(args))
	for _, e := range args {
		n := e
		switch e.(type) {
		case class.Time:
			v := e.(class.Time)
			if sqliteOrTaos {
				n = v.UnixMill()
			} else {
				n = v.Time
			}
		case time.Time:
			v := e.(time.Time)
			if sqliteOrTaos {
				n = v.UnixMilli()
			}
		case string, class.String:
			// taos未对其中的'字符转义, 但在insert中转义了？todo
			if sqlconst.IsTaos(driver) {
				var ee string
				if ev, ok := e.(string); ok {
					ee = ev
				} else {
					ee = e.(class.String).String
				}
				if strings.Contains(ee, "'") {
					n = strings.ReplaceAll(ee, "'", "''")
				}
			}
		}
		new_args = append(new_args, n)
	}
	return new_args
}

func handlePlaceholderInWhere(driver string, pred any, args ...any) any {
	_, ok := pred.(string)
	if ok && sqlconst.IsTaos(driver) && len(args) > 0 {
		noString := true
		for _, e := range args {
			if _, ok := e.(string); ok {
				noString = false
				break
			}
		}
		if noString {
			return pred
		}
		p := pred.(string)
		index := 0
		pa := ""
		ps := strings.Split(p, "?")
		for _, e := range ps {
			if len(args) <= index {
				break
			}
			if _, ok := args[index].(string); ok {
				pa += e + "'?'"
			} else {
				pa += e + "?"
			}
			index++
		}
		return pa
	}
	return pred
}
