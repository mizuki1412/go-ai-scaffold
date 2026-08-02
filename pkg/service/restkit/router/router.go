package router

import (
	"embed"
	"io"
	"mime"
	"net/http"
	"path"

	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/openapi"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

// Router router的抽象
type Router struct {
	Proxy      *gin.Engine
	ProxyGroup *gin.RouterGroup
	Openapi    *openapi.Builder
	// openapiMulti A12: 单次路由注册可能产生多个 Builder（GetPost 注册 GET+POST），
	// .Api() 需对全部 Builder 同步应用 Functional Options，否则 GET 缺失元数据。
	// 每次新路由注册（Get/Post/Put/Delete/GetPost）开始时重置。
	openapiMulti []*openapi.Builder
}
type Handler func(ctx *context.Context)

// handlerTransOne 将抽象 Handler 转为 gin.HandlerFunc。
// 独立函数避免 handlerTrans 中闭包共享循环变量（Go<1.22 下的陷阱），同时便于复用。
func handlerTransOne(handler Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler(&context.Context{
			Proxy:    ctx,
			Request:  ctx.Request,
			Response: ctx.Writer,
		})
	}
}

// handlerTrans A1: 复用 handlerTransOne，避免在循环中重新构造闭包，
// 消除潜在的循环变量捕获陷阱（即便 Go 1.22+ 已修复，复用更清晰且零成本）。
func handlerTrans(handlers ...Handler) []gin.HandlerFunc {
	list := make([]gin.HandlerFunc, len(handlers))
	for i, v := range handlers {
		list[i] = handlerTransOne(v)
	}
	return list
}

func (router *Router) Group(path string, handlers ...Handler) *Router {
	r0 := &Router{
		Proxy:      router.Proxy,
		ProxyGroup: router.ProxyGroup.Group(path),
		Openapi:    router.Openapi,
	}
	if len(handlers) > 0 {
		r0.Use(handlers...)
	}
	return r0
}

func (router *Router) Use(handlers ...Handler) *Router {
	for _, v := range handlers {
		router.ProxyGroup.Use(handlerTransOne(v))
	}
	return router
}

// Post 此处handle不能当成是use
func (router *Router) Post(path string, handlers ...Handler) *Router {
	router.ProxyGroup.POST(path, handlerTrans(handlers...)...)
	router.resetAndBuild(path, "post")
	return router
}
func (router *Router) Get(path string, handlers ...Handler) *Router {
	router.ProxyGroup.GET(path, handlerTrans(handlers...)...)
	router.resetAndBuild(path, "get")
	return router
}
func (router *Router) Put(path string, handlers ...Handler) *Router {
	router.ProxyGroup.PUT(path, handlerTrans(handlers...)...)
	router.resetAndBuild(path, "put")
	return router
}
func (router *Router) Delete(path string, handlers ...Handler) *Router {
	router.ProxyGroup.DELETE(path, handlerTrans(handlers...)...)
	router.resetAndBuild(path, "delete")
	return router
}
func (router *Router) getIgnoreOpenapi(path string, handlers ...Handler) *Router {
	router.ProxyGroup.GET(path, handlerTrans(handlers...)...)
	return router
}

// GetPost A12: 同时注册 GET 和 POST 路由，并创建两个 Builder，
// 使后续 .Api() 能同步装饰两个 method，避免 GET 在 OpenAPI 文档中缺失元数据。
func (router *Router) GetPost(path string, handlers ...Handler) *Router {
	h := handlerTrans(handlers...)
	router.ProxyGroup.GET(path, h...)
	router.ProxyGroup.POST(path, h...)
	// 重置 multi，连续追加 get 与 post 两个 Builder
	router.openapiMulti = router.openapiMulti[:0]
	router.appendBuilder(path, "get")
	router.appendBuilder(path, "post")
	return router
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	router.Proxy.ServeHTTP(w, req)
}

// openapiFullPath 拼接 group base path 与路由 path，去重中间斜杠。
func (router *Router) openapiFullPath(path string) string {
	base := router.ProxyGroup.BasePath()
	if base != "" && base[len(base)-1] == '/' && path != "" && path[0] == '/' {
		path = path[1:]
	}
	return base + path
}

// resetAndBuild 重置 multi 后注册单个 method（Get/Post/Put/Delete 通用路径）。
func (router *Router) resetAndBuild(path string, method string) {
	router.openapiMulti = router.openapiMulti[:0]
	router.appendBuilder(path, method)
}

// appendBuilder 创建 Builder 并追加到 multi（不重置，供 GetPost 连续追加）。
func (router *Router) appendBuilder(path string, method string) {
	b := openapi.NewBuilder(router.openapiFullPath(path), method)
	router.Openapi = b
	router.openapiMulti = append(router.openapiMulti, b)
}

// Api 用Functional Options的方式构建openapi参数。
// A12: 对 multi 列表中所有 Builder 同步应用 options（GetPost 场景同时装饰 GET/POST）。
func (router *Router) Api(options ...func(opt *openapi.Builder)) *Router {
	if router.Openapi == nil {
		panic(exception.New("please init openapi first"))
	}
	for _, b := range router.openapiMulti {
		for _, option := range options {
			option(b)
		}
	}
	return router
}

// EmbedHtmlHandle 注意path pattern中加入{path:path}
// url中path的路径前缀需要和root一致
//
// B8: 使用 io.ReadAll 一次性读取 + defer Close，修复文件句柄泄露与 1KB 循环低效读取。
// B9: 无扩展名时默认 text/html，避免浏览器误判 MIME。
func EmbedHtmlHandle(fs embed.FS, root string) func(c *context.Context) {
	return func(c *context.Context) {
		// 解析访问路径
		var assetPath string
		pathName := c.Proxy.Param("action")
		if root == "./knife-ui" {
			if pathName == "" || pathName == "/" {
				pathName = "doc.html"
			} else {
				pathName = "/webjars" + pathName
			}
		} else {
			if pathName == "" || pathName == "/" {
				pathName = "index.html"
			}
		}
		assetPath = path.Join(root, pathName)
		assets, err := fs.Open(assetPath)
		if err != nil {
			c.Proxy.Status(http.StatusBadRequest)
			_, _ = c.Proxy.Writer.Write([]byte(err.Error()))
			return
		}
		// B8: 必须关闭文件句柄，避免 embed.FS 句柄泄露
		defer func() { _ = assets.Close() }()
		data, err := io.ReadAll(assets)
		if err != nil {
			c.Proxy.Status(http.StatusInternalServerError)
			_, _ = c.Proxy.Writer.Write([]byte(err.Error()))
			return
		}
		// B9: mime 判断
		contentType := mime.TypeByExtension(path.Ext(pathName))
		if contentType == "" {
			// 无扩展名或未识别：默认 text/html; charset=utf-8，
			// 避免 render.Data 空 ContentType 时浏览器按二进制下载
			contentType = "text/html; charset=utf-8"
		}
		c.Proxy.Render(http.StatusOK, render.Data{Data: data, ContentType: contentType})
	}
}

func (router *Router) RegisterSwagger() {
	router.getIgnoreOpenapi("/v3/api-docs", func(c *context.Context) {
		c.Proxy.JSON(http.StatusOK, openapi.Doc.ReadDoc())
	})
	router.getIgnoreOpenapi("/v3/api-docs/swagger-config", func(c *context.Context) {
		c.Proxy.JSON(http.StatusOK, openapi.Doc.SwaggerConfig())
	})
	// 第二个path表示匹配路径
	router.getIgnoreOpenapi("/swagger/*action", EmbedHtmlHandle(openapi.UiAssets, "./swagger-ui"))
	router.getIgnoreOpenapi("/swagger", EmbedHtmlHandle(openapi.UiAssets, "./swagger-ui"))
	router.getIgnoreOpenapi("/doc.html", EmbedHtmlHandle(openapi.KUiAssets, "./knife-ui"))
	router.getIgnoreOpenapi("/webjars/*action", EmbedHtmlHandle(openapi.KUiAssets, "./knife-ui"))
}
