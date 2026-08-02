package timekit

import (
	"github.com/example/go-ai-scaffold/pkg/cli/configkey"
	"github.com/example/go-ai-scaffold/pkg/service/configkit"
	"time"
)

func GetLocation() *time.Location {
	// os.Setenv("TZ", "UTC") 环境变量的方式只能生效一次
	// time.Local 可能受实际运行环境影响
	loc, _ := time.LoadLocation(configkit.GetString(configkey.TimeLocation))
	return loc
}
