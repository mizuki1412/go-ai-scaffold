package cmd

import (
	"github.com/spf13/cobra"
)

// MarkdownDocCMD 依赖未迁移的 tool-local/markdown，已暂时停用。
func MarkdownDocCMD(title string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mdoc",
		Short: `gen readme md doc (disabled: tool-local/markdown not migrated)`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	cmd.Flags().String("dest", "", "生成目标路径 /xx/xx")
	_ = cmd.MarkFlagRequired("dest")
	return cmd
}
