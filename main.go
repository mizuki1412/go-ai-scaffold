package main

import (
	"github.com/example/go-ai-scaffold/mod/user"
	"github.com/example/go-ai-scaffold/pkg/cli"
	"github.com/example/go-ai-scaffold/pkg/service/restkit"
	"github.com/spf13/cobra"
)

func main() {
	cli.RootCMD(&cobra.Command{
		Use: "main",
		Run: func(cmd *cobra.Command, args []string) {
			restkit.AddActions(user.All()...)
			_ = restkit.Run()
		},
	})
	cli.AddChildCMDWithoutConfig(cli.VersionCMD())
	cli.Execute()
}
