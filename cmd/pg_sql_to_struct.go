package cmd

import (
	"github.com/spf13/cobra"
)

// PGSqlToStructCMD 依赖未迁移的 tool-local/pgsql，已暂时停用。
func PGSqlToStructCMD(sqlFile, destFile string) *cobra.Command {
	if sqlFile == "" {
		sqlFile = "/Users/ycj/Downloads/init.sql"
	}
	if destFile == "" {
		destFile = "/Users/ycj/Downloads/init.go"
	}
	return &cobra.Command{
		Use:   "pgss",
		Short: "pg sql to struct (disabled: tool-local/pgsql not migrated)",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
}
