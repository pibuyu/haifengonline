package captcha

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers"
	receive "haifengonline/interaction/receive/users"
	"haifengonline/logic/captcha"
)

type Controllers struct {
	controllers.BaseControllers
}

func (c *Controllers) GetCaptcha(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetCaptchaStruct)); err == nil {
		//global.Logger.Infof("使用id:%s 请求了captcha:", rec.CaptchaId)
		results, err := captcha.GetCaptcha(rec.CaptchaId)
		c.Response(ctx, results, err)
	} else {

	}
}
