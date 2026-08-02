package cmd

import (
	"github.com/spf13/cobra"
)

// FrontDaoCMDNext 依赖未迁移的 tool-local/frontdao，已暂时停用。
func FrontDaoCMDNext(url string) *cobra.Command {
	return &cobra.Command{
		Use:   "genDao",
		Short: "generate front dao (disabled: tool-local/frontdao not migrated)",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
}
