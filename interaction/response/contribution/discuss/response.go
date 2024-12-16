package discuss

import (
	articleComments "haifengonline/models/contribution/article/comments"
	"haifengonline/models/contribution/video/barrage"
	videoComments "haifengonline/models/contribution/video/comments"
	"haifengonline/utils/conversion"
	"time"
)

type GetDiscussVideoListItem struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	Photo         string    `json:"photo"`
	Comment       string    `json:"comment"`
	Cover         string    `json:"cover"`
	Title         string    `json:"title"`
	CreatedAt     time.Time `json:"created_at"`
	VideoId       uint      `json:"videoId"`
	CommentID     uint      `json:"comment_id" gorm:"column:comment_id"`
	CommentUserID uint      `json:"comment_user_id" gorm:"column:comment_user_id"`
}

type GetDiscussVideoListStruct []GetDiscussVideoListItem

func GetDiscussVideoListResponse(cml *videoComments.CommentList) interface{} {
	//判空
	if cml == nil {
		return make(GetDiscussVideoListStruct, 0)
	}
	list := make(GetDiscussVideoListStruct, 0)
	for _, v := range *cml {
		photo, _ := conversion.FormattingJsonSrc(v.UserInfo.Photo)
		cover, _ := conversion.FormattingJsonSrc(v.VideoInfo.Cover)
		list = append(list, GetDiscussVideoListItem{
			ID:            v.ID,
			Username:      v.UserInfo.Username,
			Photo:         photo,
			Comment:       v.Context,
			Cover:         cover,
			Title:         v.VideoInfo.Title,
			CreatedAt:     v.CreatedAt,
			VideoId:       v.VideoID,
			CommentID:     v.CommentID,
			CommentUserID: v.CommentUserID,
		})
	}
	return list
}

type GetDiscussArticleListItem struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Photo     string    `json:"photo"`
	Comment   string    `json:"comment"`
	Cover     string    `json:"cover"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type GetDiscussArticleListStruct []GetDiscussArticleListItem

func GetDiscussArticleListResponse(cml *articleComments.CommentList) interface{} {
	//cml的判空逻辑
	if cml == nil {
		return make(GetDiscussArticleListStruct, 0)
	}
	list := make(GetDiscussArticleListStruct, 0)
	for _, v := range *cml {
		photo, _ := conversion.FormattingJsonSrc(v.UserInfo.Photo)
		cover, _ := conversion.FormattingJsonSrc(v.ArticleInfo.Cover)
		list = append(list, GetDiscussArticleListItem{
			ID:        v.ID,
			Username:  v.UserInfo.Username,
			Photo:     photo,
			Comment:   v.Context,
			Cover:     cover,
			Title:     v.ArticleInfo.Title,
			CreatedAt: v.CreatedAt,
		})
	}
	return list
}

type GetDiscussBarrageListItem struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Photo     string    `json:"photo"`
	Comment   string    `json:"comment"`
	Cover     string    `json:"cover"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type GetDiscussBarrageListStruct []GetDiscussBarrageListItem

func GetDiscussBarrageListResponse(cml *barrage.BarragesList) interface{} {
	if cml == nil {
		return make(GetDiscussBarrageListStruct, 0)
	}
	list := make(GetDiscussBarrageListStruct, 0)
	for _, v := range *cml {
		photo, _ := conversion.FormattingJsonSrc(v.UserInfo.Photo)
		cover, _ := conversion.FormattingJsonSrc(v.VideoInfo.Cover)
		list = append(list, GetDiscussBarrageListItem{
			ID:        v.ID,
			Username:  v.UserInfo.Username,
			Photo:     photo,
			Comment:   v.Text,
			Cover:     cover,
			Title:     v.VideoInfo.Title,
			CreatedAt: v.CreatedAt,
		})
	}
	return list
}
