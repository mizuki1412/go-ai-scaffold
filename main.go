package main

import (
	"github.com/example/go-ai-scaffold/cmd"
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
	c1 := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	c1.Flags().String("test", "", "")
	c1.Flags().String("test1", "", "")
	cli.AddChildCMD(c1)

	cli.AddChildCMD(cmd.FrontDaoCMDNext("http://localhost:10000/v3/api-docs"))
	cli.Execute()
}
