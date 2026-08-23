package cli

import (
	"runtime"

	"github.com/spf13/cobra"
)

// 版本信息，构建时通过 -ldflags 注入，例如：
//
//	go build -ldflags "-X github.com/example/go-ai-scaffold/pkg/cli.buildVersion=v1.0.0 \
//	  -X github.com/example/go-ai-scaffold/pkg/cli.buildCommit=$(git rev-parse --short HEAD) \
//	  -X github.com/example/go-ai-scaffold/pkg/cli.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	buildVersion = "dev"
	buildCommit  = "unknown"
	buildDate    = "unknown"
)

type BuildInfo struct {
	GoVersion string `json:"goVersion"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
}

// GetBuildInfo 返回构建信息（未注入时为 dev/unknown）。
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		GoVersion: runtime.Version(),
		Version:   buildVersion,
		Commit:    buildCommit,
		Date:      buildDate,
	}
}

// VersionCMD version 子命令。不依赖配置文件，注册请用 AddChildCMDWithoutConfig。
func VersionCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "打印构建版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			info := GetBuildInfo()
			cmd.Printf("version:  %s\ncommit:   %s\ndate:     %s\ngo:       %s\n",
				info.Version, info.Commit, info.Date, info.GoVersion)
		},
	}
}
