package cmd

import (
	"log"

	"github.com/example/go-ai-scaffold/pkg/library/bytekit"
	"github.com/example/go-ai-scaffold/pkg/service/configkit"
	"github.com/example/go-ai-scaffold/pkg/service/netkit"
	"github.com/panjf2000/gnet/v2"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func TCPServerCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tcp-server",
		Short: "本地tcp服务器",
		Run: func(cmd *cobra.Command, args []string) {
			server := &netkit.NetServer{
				Port: cast.ToInt32(configkit.GetString("port")),
				TrafficHandler: func(c gnet.Conn) {
					buf, _ := c.Next(-1)
					log.Println("recv：" + bytekit.Bytes2HexArray(buf))
					return
				},
			}
			server.Run()
		},
	}
	cmd.Flags().String("port", "", "端口")
	_ = cmd.MarkFlagRequired("port")
	return cmd
}

// P2 修复：删除死代码 tcpServer/handleClient——
// 两函数无任何调用方，且 ResolveTCPAddr/ListenTCP 错误被忽略（nil deref）、
// Accept 错误 continue 忙转；实际 TCP 服务已由 netkit（gnet）实现。
