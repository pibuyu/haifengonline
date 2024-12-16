package callback

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers"
	receive "haifengonline/interaction/receive/callback"
	"haifengonline/logic/callback"
)

type Controllers struct {
	controllers.BaseControllers
}

// AliyunTranscodingMedia 阿里云媒体转码成功回调
func (c *Controllers) AliyunTranscodingMedia(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.AliyunMediaCallback[receive.AliyunTranscodingMediaStruct])); err == nil {
		results, err := callback.AliyunTranscodingMedia(rec)
		c.Response(ctx, results, err)
	}
}
