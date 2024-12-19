package contribution

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"haifengonline/controllers"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/contribution/video"
	"haifengonline/logic/contribution"
	"haifengonline/models/contribution/video/comments"
	"haifengonline/quartzImpl/video"
	"haifengonline/utils/response"
	"haifengonline/utils/validator"
	"strconv"
	"time"
)

type Controllers struct {
	controllers.BaseControllers
}

// CreateVideoContribution 发布视频和编辑视频
func (c Controllers) CreateVideoContribution(ctx *gin.Context) {
	uid := ctx.GetUint("uid")

	if rec, err := controllers.ShouldBind(ctx, new(receive.CreateVideoContributionReceiveStruct)); err == nil {
		results, err := contribution.CreateVideoContribution(rec, uid)
		if err != nil {
			global.Logger.Errorf("保存视频信息失败：%v", err)
		}

		//todo:定时发布视频
		if rec.DateTime != "" {
			//一定要指定时区，不然时间会不一致，算出来的结果非常离谱！！！！
			location, _ := time.LoadLocation("Local")
			targetTime, err := time.ParseInLocation("2006-01-02 15:04:05", rec.DateTime, location)
			if err != nil {
				global.Logger.Errorf("解析时间出错，传递过来的dateTime为%v", rec.DateTime)
			}
			//err = video.PublishVideo(targetTime.Sub(time.Now()), results.(uint), context.Background())
			err = video.PublishVideoOnSchedule(targetTime, results.(uint))
		}
		c.Response(ctx, results, err)
	}
}

// UpdateVideoContribution 编辑视频
func (c Controllers) UpdateVideoContribution(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.UpdateVideoContributionReceiveStruct)); err == nil {
		results, err := contribution.UpdateVideoContribution(rec, uid)
		c.Response(ctx, results, err)
	}
}

// GetVideoContributionByID  根据id获取视频信息
// 每次播放视频的时候请求视频信息，都应该使视频热度（可以理解为播放量）+1
func (c Controllers) GetVideoContributionByID(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetVideoContributionByIDReceiveStruct)); err == nil {
		results, err := contribution.GetVideoContributionByID(rec, uid)
		c.Response(ctx, results, err)
	}

}

// SendVideoBarrage  发送视频弹幕
// 步骤一：保存弹幕到数据库；步骤二：通过对话的socket通知连接到以videoId为标识的房间里的所有用户
func (c Controllers) SendVideoBarrage(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	SendVideoBarrageReceive := new(receive.SendVideoBarrageReceiveStruct)
	//使用ShouldBindBodyWith解决重复bind问题
	if err := ctx.ShouldBindBodyWith(SendVideoBarrageReceive, binding.JSON); err == nil {
		results, err := contribution.SendVideoBarrage(SendVideoBarrageReceive, uid)
		if err != nil {
			response.Error(ctx, err.Error())
			return
		}
		response.BarrageSuccess(ctx, results)
	} else {
		validator.CheckParams(ctx, err)
	}
}

// GetVideoBarrage  获取视频弹幕 (播放器）
func (c Controllers) GetVideoBarrage(ctx *gin.Context) {
	GetVideoBarrageReceive := new(receive.GetVideoBarrageReceiveStruct)
	GetVideoBarrageReceive.ID = ctx.Query("id")
	results, err := contribution.GetVideoBarrage(GetVideoBarrageReceive)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.BarrageSuccess(ctx, results)

}

// GetVideoBarrageList  获取视频弹幕展示
func (c Controllers) GetVideoBarrageList(ctx *gin.Context) {
	GetVideoBarrageReceive := new(receive.GetVideoBarrageListReceiveStruct)
	GetVideoBarrageReceive.ID = ctx.Query("id")
	//fmt.Println("获取视频弹幕请求中携带的id为", GetVideoBarrageReceive.ID)
	results, err := contribution.GetVideoBarrageList(GetVideoBarrageReceive)
	if err != nil {
		response.Error(ctx, err.Error())
		return
	}
	response.BarrageSuccess(ctx, results)
}

// VideoPostComment 视频评论
func (c Controllers) VideoPostComment(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	//params：videoId、commentContent、commentId
	if rec, err := controllers.ShouldBind(ctx, new(receive.VideosPostCommentReceiveStruct)); err == nil {
		results, err := contribution.VideoPostComment(rec, uid)
		c.Response(ctx, results, err)
	}
}

// GetVideoComment 获取视频评论
func (c Controllers) GetVideoComment(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetVideoCommentReceiveStruct)); err == nil {
		results, err := contribution.GetVideoComment(rec)
		c.Response(ctx, results, err)
	}
}

// GetVideoManagementList 创作中心获取视频稿件列表
func (c Controllers) GetVideoManagementList(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetVideoManagementListReceiveStruct)); err == nil {
		results, err := contribution.GetVideoManagementList(rec, uid)
		c.Response(ctx, results, err)
	}
}

// DeleteVideoByID 删除视频根据id
func (c Controllers) DeleteVideoByID(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.DeleteVideoByIDReceiveStruct)); err == nil {
		results, err := contribution.DeleteVideoByID(rec, uid)
		c.Response(ctx, results, err)
	}
}

func (c Controllers) DeleteVideoByPath(ctx *gin.Context) {
	//todo:这里应该加一重验证：要删除的视频是否是当前用户上传的
	if rec, err := controllers.ShouldBind(ctx, new(receive.DeleteVideoByPathReceiveStruct)); err == nil {
		results, err := contribution.DeleteVideoByPath(rec)
		c.Response(ctx, results, err)
	}
}

func (c Controllers) GetLastWatchTime(ctx *gin.Context) {
	param, _ := strconv.ParseInt(ctx.Query("id"), 10, 64)
	uid := ctx.GetUint("uid")

	results, err := contribution.GetLastWatchTime(uint(uid), uint(param))
	c.Response(ctx, results, err)
}

func (c Controllers) SendWatchTime(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.SendWatchTimeReqStruct)); err == nil {
		if err := contribution.SendWatchTime(rec, uid); err != nil {
			c.Response(ctx, "保存失败", err)
		}
		c.Response(ctx, "保存成功", err)
	}
}

// LikeVideo 给视频点赞
func (c Controllers) LikeVideo(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.LikeVideoReceiveStruct)); err == nil {
		results, err := contribution.LikeVideo(rec, uid)
		c.Response(ctx, results, err)
	}
}

func (c Controllers) LikeVideoComment(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.LikeVideoCommentReqStruct)); err == nil {
		results, err := contribution.LikeVideoComment(rec)
		c.Response(ctx, results, err)
	}
}

// GetVideoCommentCountById 根据视频id返回视频的评论总条数
func (c Controllers) GetVideoCommentCountById(ctx *gin.Context) {
	/*
		踩了几个雷：
		1、前端post方法传递过来的参数{id:"3123"}不能用ctx.Query("id")取，要构造一个json对象，然后ctx.BindJSON取参数id
		2、comments表的videoId字段为uint类型，取出id之后还要转为uint类型才能正确的进行查询
	*/
	var json struct {
		ID string `json:"id"`
	}
	if err := ctx.BindJSON(&json); err != nil {
		global.Logger.Errorf("类型转换错误：%v", err)
	}

	value, _ := strconv.Atoi(json.ID)
	videoId := uint(value)
	//fmt.Printf("从前端接收到的id为%v,类型为%v", videoId, reflect.TypeOf(videoId))
	var count int64
	err := global.Db.Model(&comments.Comment{}).Where("video_id = ?", videoId).Count(&count).Error
	//fmt.Println("查询到的评论条数为", count)
	c.Response(ctx, count, err)
}
