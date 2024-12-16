package contribution

import (
	"fmt"
	receive "haifengonline/interaction/receive/contribution/discuss"
	response "haifengonline/interaction/response/contribution/discuss"
	"haifengonline/models/contribution/article"
	articleComments "haifengonline/models/contribution/article/comments"
	"haifengonline/models/contribution/video"
	"haifengonline/models/contribution/video/barrage"
	videoComments "haifengonline/models/contribution/video/comments"
)

func GetDiscussVideoList(data *receive.GetDiscussVideoListReceiveStruct, uid uint) (results interface{}, err error) {
	//获取用户发布的视频
	videoList := new(video.VideosContributionList)
	err = videoList.GetDiscussVideoCommentList(uid)
	if err != nil {
		return nil, fmt.Errorf("查询视频相关信息失败")
	}
	//判空，用户可能没有发布过视频
	if videoList == nil || len(*videoList) == 0 {
		return response.GetDiscussVideoListResponse(nil), nil
	}

	videoIDs := make([]uint, 0)
	for _, v := range *videoList {
		videoIDs = append(videoIDs, v.ID)
	}
	//得到视频信息
	cml := new(videoComments.CommentList)
	err = cml.GetCommentListByIDs(videoIDs, data.PageInfo)

	//for _, comment := range *cml {
	//	global.Logger.Info("comment的videoInfo为", comment.VideoInfo)
	//	global.Logger.Info("comment的userInfo为", comment.UserInfo)
	//}
	//global.Logger.Info("获取视频评论返回的结果，查看是否有userInfo和videoInfo", cml)
	if err != nil {
		return nil, fmt.Errorf("查询视频评论信息失败")
	}

	return response.GetDiscussVideoListResponse(cml), nil
}

func GetDiscussArticleList(data *receive.GetDiscussArticleListReceiveStruct, uid uint) (results interface{}, err error) {
	//获取用户发布的专栏
	articleList := new(article.ArticlesContributionList)
	//这行sql返回的articleList应该是null，导致下面的articleIDs是个空集合，进而导致查询文章评论报错
	err = articleList.GetDiscussArticleCommentList(uid)

	if articleList == nil || len(*articleList) == 0 {
		return response.GetDiscussArticleListResponse(nil), nil
	}

	if err != nil {
		return nil, fmt.Errorf("查询专栏相关信息失败")
	}
	articleIDs := make([]uint, 0)
	for _, v := range *articleList {
		articleIDs = append(articleIDs, v.ID)
	}
	//global.Logger.Info("需要查询的文章id列表为：", articleIDs)
	//得到文章信息
	cml := new(articleComments.CommentList)

	err = cml.GetCommentListByIDs(articleIDs, data.PageInfo)
	if err != nil {
		return nil, fmt.Errorf("查询文章评论信息失败")
	}
	return response.GetDiscussArticleListResponse(cml), nil
}

func GetDiscussBarrageList(data *receive.GetDiscussBarrageListReceiveStruct, uid uint) (results interface{}, err error) {
	//获取用户发布的视频
	videoList := new(video.VideosContributionList)
	err = videoList.GetDiscussVideoCommentList(uid)
	if err != nil {
		return nil, fmt.Errorf("查询视频相关信息失败")
	}

	if videoList == nil || len(*videoList) == 0 {
		return response.GetDiscussBarrageListResponse(nil), nil
	}
	videoIDs := make([]uint, 0)
	for _, v := range *videoList {
		videoIDs = append(videoIDs, v.ID)
	}
	//得到视频弹幕信息
	cml := new(barrage.BarragesList)
	err = cml.GetBarrageListByIDs(videoIDs, data.PageInfo)
	if err != nil {
		return nil, fmt.Errorf("查询视频弹幕信息失败")
	}
	return response.GetDiscussBarrageListResponse(cml), nil
}
