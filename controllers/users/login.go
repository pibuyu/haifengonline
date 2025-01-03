package users

import (
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"haifengonline/controllers"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/users"
	"haifengonline/logic/users"
	"haifengonline/utils/limiter"
	"time"
)

type LoginControllers struct {
	controllers.BaseControllers
}

// WxAuthorization 微信快捷登入
func (lg LoginControllers) WxAuthorization(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.WxAuthorizationReceiveStruct)); err == nil {
		results, err := users.WxAuthorization(rec)
		lg.Response(ctx, results, err)
	}
}

// Login 登入
func (lg LoginControllers) Login(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.LoginReceiveStruct)); err == nil {
		results, err := users.Login(rec)
		lg.Response(ctx, results, err)
	}
}

// Register 注册
func (lg LoginControllers) Register(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.RegisterReceiveStruct)); err == nil {
		results, err := users.Register(rec)
		lg.Response(ctx, results, err)
	}
}

// SendEmailVerCode 获取验证码(注册)
func (lg LoginControllers) SendEmailVerCode(ctx *gin.Context) {

	if rec, err := controllers.ShouldBind(ctx, new(receive.SendEmailVerCodeReceiveStruct)); err == nil {
		//针对每个to email进行限制，对一个邮箱短时间内不允许多次请求
		//r:填充速率,b:bucket的大小
		limit := limiter.NewLimiter(rate.Every(10*time.Second), 10, rec.Email)
		if !limit.Allow() {
			lg.Response(ctx, nil, errors.New("请求过于频繁，请1分钟后再试"))
			return
		}

		results, err := users.SendEmailVerCode(rec)
		global.Logger.Infof("向邮箱：%s发送验证码", rec.Email)
		lg.Response(ctx, results, err)
	}

}

// SendEmailVerCodeByForget 获取邮箱验证码(忘记密码)
func (lg LoginControllers) SendEmailVerCodeByForget(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.SendEmailVerCodeReceiveStruct)); err == nil {
		results, err := users.SendEmailVerCodeByForget(rec)
		lg.Response(ctx, results, err)
	}
}

// Forget 找回密码
func (lg LoginControllers) Forget(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.ForgetReceiveStruct)); err == nil {
		results, err := users.Forget(rec)
		lg.Response(ctx, results, err)
	}
}
