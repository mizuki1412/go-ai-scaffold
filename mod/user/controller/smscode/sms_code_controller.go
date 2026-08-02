package smscode

import (
	context2 "context"
	"crypto/rand"
	"time"

	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/library/regexkit"
	"github.com/example/go-ai-scaffold/pkg/service/rediskit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/openapi"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

func Init(router *router.Router) {
	tag := "user:用户模块"
	r := router.Group("/user")
	r.Post("/getVerifyCode", GetVerifyCode).Api(openapi.Tag(tag), openapi.Summary("短信验证码获取"), openapi.ReqParam(getParams{}),
		openapi.Security(nil))
}

type getParams struct {
	Phone string `comment:"手机号" validate:"required" trim:"true"`
}

// smsCodeExpire 验证码有效期
const smsCodeExpire = 10 * time.Minute

// smsSendInterval 同一手机号两次发送的最小间隔，防止刷接口
const smsSendInterval = 60 * time.Second

func GetVerifyCode(ctx *context.Context) {
	params := getParams{}
	ctx.BindForm(&params)
	if !regexkit.IsPhone(params.Phone) {
		panic(exception.New("手机号码格式错误"))
	}
	// B18: 频率限制 — 同一手机号 60s 内只能发一次
	limitKey := rediskit.GetKeyWithPrefix("sms:limit:" + params.Phone)
	if rediskit.Get(context2.Background(), limitKey, "") != "" {
		panic(exception.New("发送过于频繁，请稍后再试"))
	}
	// B17: 用 crypto/rand 生成 4 位数字验证码，避免 math/rand 可预测
	sms := genSmsCode(4)
	rediskit.Set(context2.Background(), rediskit.GetKeyWithPrefix("sms:"+params.Phone), sms, smsCodeExpire)
	rediskit.Set(context2.Background(), limitKey, "1", smsSendInterval)
	// 未迁移模块，按需补充：alismskit.Send(alismskit.SendParams{Phone: params.Phone, Data: map[string]any{"code": sms}})
	ctx.JsonSuccess()
}

// genSmsCode 用 crypto/rand 生成 n 位数字验证码字符串。
// 每位独立采样 0-9，剔除 [250,255] 避免模偏置；返回值长度固定为 n，可含前导 0。
func genSmsCode(n int) string {
	buf := make([]byte, n)
	// 一次读取 n 字节随机数；失败则退化为时间戳兜底，保证流程不中断
	if _, err := rand.Read(buf); err != nil {
		now := time.Now().UnixNano()
		for i := range buf {
			buf[i] = byte('0' + (int(now>>(uint(i)*4)) % 10))
		}
		return string(buf)
	}
	for i := range buf {
		for buf[i] >= 250 {
			rand.Read(buf[i : i+1])
		}
		buf[i] = byte('0' + buf[i]%10)
	}
	return string(buf)
}
