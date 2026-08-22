package restkit

import (
	ctx "context"
	"errors"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/go-ai-scaffold/pkg/cli/configkey"
	"github.com/example/go-ai-scaffold/pkg/service/configkit"
	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/middleware"
	router2 "github.com/example/go-ai-scaffold/pkg/service/restkit/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var router *router2.Router
var server *http.Server

func defaultEngine() {
	if !configkit.GetBool(configkey.ProfileDev) {
		gin.SetMode(gin.ReleaseMode)
	}
	router = &router2.Router{
		Proxy: gin.New(),
	}
	// add base path
	base := configkit.GetString(configkey.RestServerBase)
	if base != "" {
		if base[0] != '/' {
			base = "/" + base
		}
		if base[len(base)-1] == '/' {
			base = base[:len(base)-1]
		}
		router.ProxyGroup = router.Proxy.Group(base)
	} else {
		router.ProxyGroup = &router.Proxy.RouterGroup
	}
	router.Use(middleware.Log())
	router.Use(middleware.Cors())
	router.Use(middleware.Recover())
	if configkit.GetBool(configkey.RestPPROF) {
		// P2 修复：原开关为空实现（死开关），接通 pprof 挂载。
		// pprof 端点暴露运行时内部信息（堆、goroutine 栈等），仅开发/内网诊断时开启。
		registerPprof(router.Proxy)
	}
	// max request size todo
	//router.Proxy.Use(iris.LimitRequestBodySize(int64(configkit.GetInt(configkey.RestRequestBodySize, 100)) << 20))
	// 其他错误如404，
	//router.OnError(middleware.Cors())
}

// registerPprof 在 gin 引擎上挂载 net/http/pprof 端点（/debug/pprof/*），
// 不走 Router 路由封装（openapi 元数据对诊断端点无意义），也不受 base path 影响。
func registerPprof(engine *gin.Engine) {
	g := engine.Group("/debug/pprof")
	g.GET("", gin.WrapF(pprof.Index))
	g.GET("/", gin.WrapF(pprof.Index))
	g.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	g.GET("/profile", gin.WrapF(pprof.Profile))
	g.POST("/symbol", gin.WrapF(pprof.Symbol))
	g.GET("/symbol", gin.WrapF(pprof.Symbol))
	g.GET("/trace", gin.WrapF(pprof.Trace))
	g.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
	g.GET("/block", gin.WrapH(pprof.Handler("block")))
	g.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	g.GET("/heap", gin.WrapH(pprof.Handler("heap")))
	g.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	g.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}

// newHTTPServer 构造带超时配置的 http.Server。
// P1 修复：原来四项超时全缺，慢客户端可无限占用连接（slowloris 攻击面）。
// WriteTimeout 默认 0（不限制）：脚手架含 SSE 长连接场景（ssehelper），
// 写超时会把事件流截断；其余三项给出安全默认值，均可经配置覆盖。
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       time.Duration(configkit.GetInt(configkey.RestReadTimeout, 60)) * time.Second,
		ReadHeaderTimeout: time.Duration(configkit.GetInt(configkey.RestReadHeaderTimeout, 10)) * time.Second,
		IdleTimeout:       time.Duration(configkit.GetInt(configkey.RestIdleTimeout, 120)) * time.Second,
	}
	if writeTimeout := configkit.GetInt(configkey.RestWriteTimeout, 0); writeTimeout > 0 {
		srv.WriteTimeout = time.Duration(writeTimeout) * time.Second
	}
	return srv
}

func Run(listeners ...net.Listener) error {
	if router == nil {
		defaultEngine()
	}
	port := configkit.GetString(configkey.RestServerPort)
	router.RegisterSwagger()
	if len(listeners) == 0 {
		server = newHTTPServer(":"+port, router)
	} else {
		server = newHTTPServer("", router)
		port = cast.ToString(listeners[0].Addr().(*net.TCPAddr).Port)
	}
	go func() {
		logkit.Info("Listening and serving HTTP on " + port)
		// P2 修复：serve 失败（端口占用等）时进程已不可用。
		// 原裸 panic 在 goroutine 中表现为一屏无上下文堆栈后退出，
		// 改为 Fatal：日志给出明确原因后以退出码 1 结束
		if len(listeners) == 0 {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logkit.Fatal("HTTP server serve error", "err", err.Error())
			}
		} else {
			if err := server.Serve(listeners[0]); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logkit.Fatal("HTTP server serve error", "err", err.Error())
			}
		}
	}()
	// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-without-context/server.go
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)
	<-quit
	logkit.Info("Shutting down server...")

	ctxt, cancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
	defer cancel()
	// 执行自定义关机逻辑
	CustomShutdownLogic(ctxt)
	if err := server.Shutdown(ctxt); err != nil {
		logkit.Error(err.Error())
		return err
	}
	return nil
}

// Shutdown 主动关停（与 Run 的信号触发关闭等价）。
// P2 修复：原用 context.Background 无超时，连接不释放时 Shutdown 永久阻塞；
// 与 Run 一致给 5s 优雅关闭窗口，超时后返回并由调用方决定后续（强杀或告警）。
func Shutdown() {
	if server == nil {
		return
	}
	logkit.Info("Shutting down server...")
	ctxt, cancel := ctx.WithTimeout(ctx.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctxt); err != nil {
		logkit.Error(err.Error())
	}
}

var CustomShutdownLogic = func(ctx ctx.Context) {}

// AddActions 导入业务模块，其中的路由和中间件
func AddActions(actionInits ...func(r *router2.Router)) {
	if router == nil {
		defaultEngine()
	}
	for _, action := range actionInits {
		action(router)
	}
}

func GetRouter() *router2.Router {
	if router == nil {
		defaultEngine()
	}
	return router
}
