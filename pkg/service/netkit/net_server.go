package netkit

import (
	"fmt"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/panjf2000/gnet/v2"
	"github.com/spf13/cast"
)

type NetServer struct {
	gnet.BuiltinEventEngine
	eng         gnet.Engine
	ProtoSchema string `comment:"tcp/udp"`
	Port        int32

	TrafficHandler func(c gnet.Conn)
}

func (th *NetServer) OnBoot(eng gnet.Engine) gnet.Action {
	th.eng = eng
	logkit.Info("net server is listening on " + cast.ToString(th.Port))
	return gnet.None
}

func (th *NetServer) OnTraffic(c gnet.Conn) gnet.Action {
	if th.TrafficHandler != nil {
		th.TrafficHandler(c)
	}
	return gnet.None
}

func (th *NetServer) Run() {
	if th.ProtoSchema == "" {
		th.ProtoSchema = "tcp"
	}
	err := gnet.Run(
		th,
		fmt.Sprintf("%s://:%d", th.ProtoSchema, th.Port),
		gnet.WithMulticore(true),
		gnet.WithReusePort(true))
	if err != nil {
		panic(exception.New(err.Error()))
	}
}
