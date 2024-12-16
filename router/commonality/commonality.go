package commonality

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/commonality"
	"haifengonline/middlewares"
)

func (r *RouterGroup) InitRouter(Router *gin.RouterGroup) {
	commonalityControllers := new(commonality.Controllers)
	routers := Router.Group("commonality").Use()
	{
		routers.POST("/ossSTS", commonalityControllers.OssSTS)
		routers.POST("/upload", commonalityControllers.Upload)
		routers.POST("/UploadSlice", commonalityControllers.UploadSlice)
		routers.POST("/uploadCheck", commonalityControllers.UploadCheck)
		routers.POST("/uploadMerge", commonalityControllers.UploadMerge)
		routers.POST("/uploadingMethod", commonalityControllers.UploadingMethod)
		routers.POST("/uploadingDir", commonalityControllers.UploadingDir)
		routers.POST("/getFullPathOfImage", commonalityControllers.GetFullPathOfImage)
		routers.POST("/registerMedia", commonalityControllers.RegisterMedia)

	}
	//非必须登入
	contributionRouterNotNecessary := Router.Group("commonality").Use(middlewares.VerificationTokenNotNecessary())
	{
		contributionRouterNotNecessary.POST("/search", commonalityControllers.Search)
	}
}
