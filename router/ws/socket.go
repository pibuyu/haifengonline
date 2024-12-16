package ws

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/contribution"
	"haifengonline/controllers/live"
	"haifengonline/controllers/users"
	"haifengonline/middlewares"
)

func (r *RouterGroup) InitSocketRouter(Router *gin.RouterGroup) {
	socketRouter := Router.Group("ws").Use(middlewares.VerificationTokenAsSocket())
	{
		userControllers := new(users.UserControllers)
		liveControllers := new(live.LivesControllers)
		contributionControllers := new(contribution.Controllers)
		socketRouter.GET("/noticeSocket", userControllers.NoticeSocket)
		socketRouter.GET("/chatSocket", userControllers.ChatSocket)
		socketRouter.GET("/chatUserSocket", userControllers.ChatByUserSocket)
		socketRouter.GET("/liveSocket", liveControllers.LiveSocket)
		socketRouter.GET("/videoSocket", contributionControllers.VideoSocket)
	}
}
