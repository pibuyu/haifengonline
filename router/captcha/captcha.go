package captcha

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/captcha"
)

func (r *RouterGroup) InitRouter(Router *gin.RouterGroup) {
	//global.Logger.Infoln("将captcha注册到router里去")
	captchaControllers := new(captcha.Controllers)
	router := Router.Group("captcha").Use()
	{
		router.POST("/getCaptcha", captchaControllers.GetCaptcha)
	}
}
