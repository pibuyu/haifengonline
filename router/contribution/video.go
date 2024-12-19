package contribution

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/contribution"
	"haifengonline/middlewares"
)

type VideoRouter struct {
}

func (v *VideoRouter) InitVideoRouter(Router *gin.RouterGroup) {
	contributionControllers := new(contribution.Controllers)
	//不需要登入
	contributionRouterNoVerification := Router.Group("contribution").Use()
	{
		contributionRouterNoVerification.GET("/video/barrage/v3/", contributionControllers.GetVideoBarrage)
		contributionRouterNoVerification.GET("/getVideoBarrageList", contributionControllers.GetVideoBarrageList)
		contributionRouterNoVerification.POST("/getVideoComment", contributionControllers.GetVideoComment)
		contributionRouterNoVerification.POST("/getVideoCommentCountById", contributionControllers.GetVideoCommentCountById)

		//给评论点赞，先不需要登录
		contributionRouterNoVerification.POST("/likeVideoComment", contributionControllers.LikeVideoComment)
	}
	//非必须登入
	contributionRouterNotNecessary := Router.Group("contribution").Use(middlewares.VerificationTokenNotNecessary())
	{
		contributionRouterNotNecessary.POST("/getVideoContributionByID", contributionControllers.GetVideoContributionByID)
	}
	//需要登入 参数携带
	contributionRouterParameter := Router.Group("contribution").Use(middlewares.VerificationTokenAsParameter())
	{
		contributionRouterParameter.POST("/video/barrage/v3/", contributionControllers.SendVideoBarrage)

	}
	//请求头携带
	contributionRouter := Router.Group("contribution").Use(middlewares.VerificationToken())
	{
		contributionRouter.POST("/createVideoContribution", contributionControllers.CreateVideoContribution)
		contributionRouter.POST("/updateVideoContribution", contributionControllers.UpdateVideoContribution)
		contributionRouter.POST("/deleteVideoByID", contributionControllers.DeleteVideoByID)
		contributionRouter.POST("/videoPostComment", contributionControllers.VideoPostComment)
		contributionRouter.POST("/getVideoManagementList", contributionControllers.GetVideoManagementList)
		contributionRouter.POST("/likeVideo", contributionControllers.LikeVideo)
		contributionRouter.POST("/deleteVideoByPath", contributionControllers.DeleteVideoByPath)
		contributionRouter.GET("/getLastWatchTime", contributionControllers.GetLastWatchTime)
		contributionRouter.POST("/sendWatchTime", contributionControllers.SendWatchTime)
	}

}
