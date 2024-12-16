package home

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers"
	receive "haifengonline/interaction/receive/home"
	"haifengonline/logic/home"
)

type Controllers struct {
	controllers.BaseControllers
}

// GetHomeInfo 获取主页信息
func (c Controllers) GetHomeInfo(ctx *gin.Context) {
	//参数有page、size、keyword
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetHomeInfoReceiveStruct)); err == nil {
		results, err := home.GetHomeInfo(rec)
		c.Response(ctx, results, err)
	}
}

func (c Controllers) SubmitBug(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.SubmitBugReceiveStruct)); err == nil {
		results, err := home.SubmitBug(rec)
		c.Response(ctx, results, err)
	}
}
