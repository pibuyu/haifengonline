package live

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/live"
	"haifengonline/middlewares"
)

type LivesRouter struct {
}

func (s *LivesRouter) InitLiveRouter(Router *gin.RouterGroup) {
	liveRouter := Router.Group("live").Use(middlewares.VerificationToken())
	{
		liveControllers := new(live.LivesControllers)
		liveRouter.POST("/getLiveRoom", liveControllers.GetLiveRoom)
		liveRouter.POST("/getLiveRoomInfo", liveControllers.GetLiveRoomInfo)
		liveRouter.POST("/getBeLiveList", liveControllers.GetBeLiveList)
	}
}
