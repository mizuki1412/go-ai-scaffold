package user

import (
	"github.com/example/go-ai-scaffold/mod/user/model"
	"github.com/example/go-ai-scaffold/mod/user/service"
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/service/jwtkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
)

type loginByUsernameParam struct {
	Username string `comment:"用户名" validate:"required"`
	Pwd      string `validate:"required"`
}

type ResLogin struct {
	User  *model.User `json:"user"`
	Token string      `json:"token"`
}

func LoginByUsername(ctx *context.Context) {
	params := loginByUsernameParam{}
	ctx.BindForm(&params)
	user := service.Login(params.Username, "", params.Pwd)
	claim := jwtkit.New(user.Id)
	token := claim.Token()
	ctx.SetJwtCookie(claim, token)
	ret := ResLogin{
		User:  user,
		Token: token,
	}
	if AdditionLoginFunc != nil {
		AdditionLoginFunc(ctx, ret)
	}
	ctx.JsonSuccess(ret)
}

type loginParam struct {
	Username string `comment:"用户名"`
	Phone    string `comment:"手机号"`
	Pwd      string `validate:"required"`
}

// Login 通用登录（用户名或手机号）
func Login(ctx *context.Context) {
	params := loginParam{}
	ctx.BindForm(&params)
	user := service.Login(params.Username, params.Phone, params.Pwd)
	claim := jwtkit.New(user.Id)
	token := claim.Token()
	ret := ResLogin{
		User:  user,
		Token: token,
	}
	ctx.SetJwtCookie(claim, token)
	if AdditionLoginFunc != nil {
		AdditionLoginFunc(ctx, ret)
	}
	ctx.JsonSuccess(ret)
}

var AdditionLoginFunc func(ctx *context.Context, ret ResLogin)

var AdditionUserExFunc func(ctx *context.Context, u *model.User)

var AdditionUserInfoWithIdFunc = func(ctx *context.Context, u *model.User) {
	// 默认不支持普通用户获取其他用户信息
	panic(exception.New("无权限获取用户信息"))
}

type infoParam struct {
	Id class.Int64 `comment:"不填获取自己，并且返回的是user和token；否则只返回user"`
}

func Info(ctx *context.Context) {
	params := infoParam{}
	ctx.BindForm(&params)
	if !params.Id.Valid {
		// 获取自己的
		uid := ctx.GetJwt().IdInt64()
		user := service.GetUserById(uid)
		if user == nil {
			panic(exception.New("用户不存在"))
		}
		claim := jwtkit.New(user.Id)
		token := claim.Token()
		ret := ResLogin{
			User:  user,
			Token: token,
		}
		ctx.SetJwtCookie(claim, token)
		if AdditionUserExFunc != nil {
			AdditionUserExFunc(ctx, user)
		}
		ctx.JsonSuccess(ret)
	} else {
		user := service.GetUserById(params.Id.Int64)
		if user == nil {
			panic(exception.New("无此用户"))
		}
		if AdditionUserExFunc != nil {
			AdditionUserExFunc(ctx, user)
		}
		AdditionUserInfoWithIdFunc(ctx, user)
		ctx.JsonSuccess(user)
	}
}

func Logout(ctx *context.Context) {
	ctx.DestroyJwt()
	ctx.JsonSuccess()
}

type updatePwdParam struct {
	OldPwd string `validate:"required"`
	NewPwd string `validate:"required"`
}

func UpdatePwd(ctx *context.Context) {
	params := updatePwdParam{}
	ctx.BindForm(&params)
	uid := ctx.GetJwt().IdInt64()
	service.UpdatePwd(uid, params.OldPwd, params.NewPwd)
	ctx.JsonSuccess()
}

type updateUserInfoParam = service.UpdateUserInfoParams

func UpdateUserInfo(ctx *context.Context) {
	params := updateUserInfoParam{}
	ctx.BindForm(&params)
	uid := ctx.GetJwt().IdInt64()
	service.UpdateUserInfo(uid, params)
	ctx.JsonSuccess()
}
