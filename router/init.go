package router

import (
	"github.com/gin-gonic/gin"
	"haifengonline/middlewares"
	"haifengonline/router/callback"
	"haifengonline/router/captcha"
	"haifengonline/router/commonality"
	"haifengonline/router/contribution"
	"haifengonline/router/home"
	"haifengonline/router/live"
	usersRouter "haifengonline/router/users"
	"haifengonline/router/ws"
)

type RoutersGroup struct {
	Users        usersRouter.RouterGroup
	Live         live.RouterGroup
	Home         home.RouterGroup
	Commonality  commonality.RouterGroup
	Contribution contribution.RouterGroup
	Ws           ws.RouterGroup
	Callback     callback.RouterGroup
	Captcha      captcha.RouterGroup
}

var RoutersGroupApp = new(RoutersGroup)

func InitRouter() {
	router := gin.Default()
	router.Use(middlewares.Cors())
	PrivateGroup := router.Group("")
	PrivateGroup.Use()
	{
		//静态资源访问
		router.Static("/assets", "./assets")
		RoutersGroupApp.Users.LoginRouter.InitLoginRouter(PrivateGroup)
		RoutersGroupApp.Users.SpaceRouter.InitSpaceRouter(PrivateGroup)
		RoutersGroupApp.Ws.InitSocketRouter(PrivateGroup)
		RoutersGroupApp.Users.InitRouter(PrivateGroup)
		RoutersGroupApp.Live.InitLiveRouter(PrivateGroup)
		RoutersGroupApp.Home.InitLiveRouter(PrivateGroup)
		RoutersGroupApp.Commonality.InitRouter(PrivateGroup)
		RoutersGroupApp.Contribution.VideoRouter.InitVideoRouter(PrivateGroup)
		RoutersGroupApp.Contribution.ArticleRouter.InitArticleRouter(PrivateGroup)
		RoutersGroupApp.Contribution.DiscussRouter.InitDiscussRouter(PrivateGroup)
		RoutersGroupApp.Callback.InitRouter(PrivateGroup)
		RoutersGroupApp.Captcha.InitRouter(PrivateGroup) //注册验证码相关的router
	}

	err := router.Run(":8081")
	if err != nil {
		return
	}
}
