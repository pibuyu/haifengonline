package users

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/users"
)

type LoginRouter struct {
}

func (s *LoginRouter) InitLoginRouter(Router *gin.RouterGroup) {
	loginRouter := Router.Group("login").Use()
	{
		loginControllers := new(users.LoginControllers)
		loginRouter.POST("/wxAuthorization", loginControllers.WxAuthorization)
		loginRouter.POST("/register", loginControllers.Register)
		loginRouter.POST("/login", loginControllers.Login)
		loginRouter.POST("/sendEmailVerificationCode", loginControllers.SendEmailVerCode)
		loginRouter.POST("/sendEmailVerificationCodeByForget", loginControllers.SendEmailVerCodeByForget)
		loginRouter.POST("/forget", loginControllers.Forget)
	}
}
