package department

import (
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/mod/user/service"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/middleware"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/openapi"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

func Init(router *router.Router) {
	tag := "department:用户模块-部门管理"
	r := router.Group("/department").Use(middleware.AuthJWT())
	r.Post("/create", CreateDepartment).Api(openapi.Tag(tag), openapi.Summary("部门新增"), openapi.ReqBody(createParams{}))
	r.Post("/update", UpdateDepartment).Api(openapi.Tag(tag), openapi.Summary("部门修改"), openapi.ReqBody(updateParams{}))
	r.Post("/del", DeleteDepartment).Api(openapi.Tag(tag), openapi.Summary("部门删除"), openapi.ReqParam(delParams{}))
	r.Post("/list", ListDepartments).Api(openapi.Tag(tag), openapi.Summary("部门列表"), openapi.Response([]*model.Department{}))
}

type createParams = service.CreateDepartmentParams

func CreateDepartment(ctx *context.Context) {
	params := createParams{}
	ctx.BindForm(&params)
	service.CreateDepartment(params)
	ctx.JsonSuccess()
}

type updateParams = service.UpdateDepartmentParams

func UpdateDepartment(ctx *context.Context) {
	params := updateParams{}
	ctx.BindForm(&params)
	service.UpdateDepartment(params)
	ctx.JsonSuccess()
}

type delParams = service.DeleteDepartmentParams

func DeleteDepartment(ctx *context.Context) {
	params := delParams{}
	ctx.BindForm(&params)
	service.DeleteDepartment(params.Id)
	ctx.JsonSuccess()
}

func ListDepartments(ctx *context.Context) {
	ctx.JsonSuccess(service.ListDepartments())
}
