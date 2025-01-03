package contribution

import (
	"github.com/gin-gonic/gin"
	"haifengonline/controllers"
	receive "haifengonline/interaction/receive/contribution/article"
	"haifengonline/logic/contribution"
)

// CreateArticleContribution 发布专栏
func (c Controllers) CreateArticleContribution(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.CreateArticleContributionReceiveStruct)); err == nil {
		results, err := contribution.CreateArticleContribution(rec, uid)
		c.Response(ctx, results, err)
	}
}

// UpdateArticleContribution 更新专栏
func (c Controllers) UpdateArticleContribution(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.UpdateArticleContributionReceiveStruct)); err == nil {
		results, err := contribution.UpdateArticleContribution(rec, uid)
		c.Response(ctx, results, err)
	}
}

// DeleteArticleByID 删除专栏
func (c Controllers) DeleteArticleByID(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.DeleteArticleByIDReceiveStruct)); err == nil {
		results, err := contribution.DeleteArticleByID(rec, uid)
		c.Response(ctx, results, err)
	}
}

// GetArticleContributionList 首页查询专栏
func (c Controllers) GetArticleContributionList(ctx *gin.Context) {
	//global.Logger.Infoln("请求了文章列表")
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetArticleContributionListReceiveStruct)); err == nil {
		results, err := contribution.GetArticleContributionList(rec)
		c.Response(ctx, results, err)
	}
}

// GetArticleContributionListByUser 查询用户发布的专栏
func (c Controllers) GetArticleContributionListByUser(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetArticleContributionListByUserReceiveStruct)); err == nil {
		results, err := contribution.GetArticleContributionListByUser(rec)
		c.Response(ctx, results, err)
	}
}

// GetArticleContributionByID 查询专栏信息根据ID
func (c Controllers) GetArticleContributionByID(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetArticleContributionByIDReceiveStruct)); err == nil {
		results, err := contribution.GetArticleContributionByID(rec, uid)
		c.Response(ctx, results, err)
	}
}

// ArticlePostComment 发布评论
func (c Controllers) ArticlePostComment(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.ArticlesPostCommentReceiveStruct)); err == nil {
		results, err := contribution.ArticlePostComment(rec, uid)
		c.Response(ctx, results, err)
	}
}

// GetArticleComment 获取文章评论
func (c Controllers) GetArticleComment(ctx *gin.Context) {
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetArticleCommentReceiveStruct)); err == nil {
		results, err := contribution.GetArticleComment(rec)
		c.Response(ctx, results, err)
	}
}

// GetArticleClassificationList 获取视频分类
func (c Controllers) GetArticleClassificationList(ctx *gin.Context) {
	results, err := contribution.GetArticleClassificationList()
	c.Response(ctx, results, err)
}

// GetArticleTotalInfo 获取文章相关总和信息
func (c Controllers) GetArticleTotalInfo(ctx *gin.Context) {
	results, err := contribution.GetArticleTotalInfo()
	c.Response(ctx, results, err)
}

// GetArticleManagementList 创作中心获取专栏稿件列表
func (c Controllers) GetArticleManagementList(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetArticleManagementListReceiveStruct)); err == nil {
		results, err := contribution.GetArticleManagementList(rec, uid)
		c.Response(ctx, results, err)
	}
}

// GetColumnByClassificationId 根据专栏id获取用户创建的对应分类下的专栏
func (c Controllers) GetColumnByClassificationId(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	if rec, err := controllers.ShouldBind(ctx, new(receive.GetColumnByClassificationId)); err == nil {
		//查看浏览器发现请求携带的参数为{columnClassificationId: 9}，字段名与GetColumnByClassificationId的json名不一致，自然绑定不上
		//global.Logger.Infof("接收到的classificationId为%v", rec.ClassificationID)
		results, err := contribution.GetColumnByClassificationId(rec, uid)
		c.Response(ctx, results, err)
	}
}
