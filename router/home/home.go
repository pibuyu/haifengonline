package home

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/home"
)

type homeRouter struct {
}

func (s *homeRouter) InitLiveRouter(Router *gin.RouterGroup) {
	homeRouter := Router.Group("home")
	{
		homeControllers := new(home.Controllers)
		homeRouter.POST("/getHomeInfo", homeControllers.GetHomeInfo)
		homeRouter.POST("/submitBug", homeControllers.SubmitBug)
	}
}
