package callback

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers/callback"
)

func (r *RouterGroup) InitRouter(Router *gin.RouterGroup) {
	callbackControllers := new(callback.Controllers)
	routers := Router.Group("callback").Use()
	{
		routers.POST("/aliyunTranscodingMedia", callbackControllers.AliyunTranscodingMedia)
	}
}
