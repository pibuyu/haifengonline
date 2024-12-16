package response

import (
	"encoding/json"
	"haifengonline/models/common"
	"haifengonline/models/contribution/video"
	"haifengonline/models/contribution/video/barrage"
	comments2 "haifengonline/models/contribution/video/comments"
	"haifengonline/models/users"
	"haifengonline/utils/conversion"
	"time"
)

// Info 视频信息
type Info struct {
	ID             uint             `json:"id"`
	Uid            uint             `json:"uid" `
	Title          string           `json:"title" `
	Video          string           `json:"video"`
	Video720p      string           `json:"video_720p"`
	Video480p      string           `json:"video_480p"`
	Video360p      string           `json:"video_360p"`
	Cover          string           `json:"cover" `
	VideoDuration  int64            `json:"video_duration"`
	Label          []string         `json:"label"`
	Introduce      string           `json:"introduce"`
	Heat           int              `json:"heat"`
	BarrageNumber  int              `json:"barrageNumber"`
	Comments       commentsInfoList `json:"comments"`
	IsLike         bool             `json:"is_like"`
	IsCollect      bool             `json:"is_collect"`
	CommentsNumber int              `json:"comments_number"`
	CreatorInfo    creatorInfo      `json:"creatorInfo"`
	CreatedAt      time.Time        `json:"created_at"`
	LikeNum        int              `json:"like_num"`
}

// 创作者信息
type creatorInfo struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	Signature   string `json:"signature"`
	IsAttention bool   `json:"is_attention"`
}

// 推荐视频信息
type recommendVideo struct {
	ID            uint      `json:"id"`
	Uid           uint      `json:"uid" `
	Title         string    `json:"title" `
	Video         string    `json:"video"`
	Cover         string    `json:"cover" `
	VideoDuration int64     `json:"video_duration"`
	Label         []string  `json:"label"`
	Introduce     string    `json:"introduce"`
	Heat          int       `json:"heat"`
	BarrageNumber int       `json:"barrageNumber"`
	Username      string    `json:"username"`
	CreatedAt     time.Time `json:"created_at"`
}
type RecommendList []recommendVideo

type Response struct {
	VideoInfo     Info          `json:"videoInfo"`
	RecommendList RecommendList `json:"recommendList"`
}

func GetVideoContributionByIDResponse(vc *video.VideosContribution, recommendVideoList *video.VideosContributionList, isAttention bool, isLike bool, isCollect bool) Response {
	//处理视频主要信息
	creatorAvatar, _ := conversion.FormattingJsonSrc(vc.UserInfo.Photo)
	cover, _ := conversion.FormattingJsonSrc(vc.Cover)
	videoSrc, _ := conversion.FormattingJsonSrc(vc.Video)
	video720pSrc, _ := conversion.FormattingJsonSrc(vc.Video720p)
	video480pSrc, _ := conversion.FormattingJsonSrc(vc.Video480p)
	video360pSrc, _ := conversion.FormattingJsonSrc(vc.Video360p)
	//评论
	comments := commentsInfoList{}
	//格式化vc.Comments，存为comments对象
	for _, v := range vc.Comments {
		photo, _ := conversion.FormattingJsonSrc(v.UserInfo.Photo)
		commentUser := users.User{}
		commentUser.Find(v.CommentUserID)
		var commentIDContent comments2.Comment
		commentIDContent.Find(v.CommentID)
		comments = append(comments, &commentsInfo{
			ID:               v.ID,
			CommentID:        v.CommentID,
			CommentIDContent: commentIDContent.Context,
			CommentFirstID:   v.CommentFirstID,
			CommentUserID:    v.CommentUserID,
			CommentUserName:  commentUser.Username,
			CreatedAt:        v.CreatedAt,
			Context:          v.Context,
			Uid:              v.UserInfo.ID,
			Username:         v.UserInfo.Username,
			Photo:            photo,
		})
	}

	//现在生成的树结构是对的，但是评论区只能展示到一级子评论，无法展示更深层次的子评论
	commentsList := comments.getChildComment()

	response := Response{
		VideoInfo: Info{
			ID:             vc.ID,
			Uid:            vc.Uid,
			Title:          vc.Title,
			Video:          videoSrc,
			Video720p:      video720pSrc,
			Video480p:      video480pSrc,
			Video360p:      video360pSrc,
			Cover:          cover,
			VideoDuration:  vc.VideoDuration,
			Label:          conversion.StringConversionMap(vc.Label),
			Introduce:      vc.Introduce,
			Heat:           vc.Heat,
			BarrageNumber:  len(vc.Barrage),
			Comments:       commentsList,
			CommentsNumber: len(commentsList),
			IsLike:         isLike,
			IsCollect:      isCollect,
			CreatorInfo: creatorInfo{
				ID:          vc.UserInfo.ID,
				Username:    vc.UserInfo.Username,
				Avatar:      creatorAvatar,
				Signature:   vc.UserInfo.Signature,
				IsAttention: isAttention,
			},
			CreatedAt: vc.CreatedAt,
			LikeNum:   len(vc.Likes),
		},
	}
	//处理推荐视频
	rl := make(RecommendList, 0)
	for _, lk := range *recommendVideoList {
		cover, _ := conversion.FormattingJsonSrc(lk.Cover)
		videoSrc, _ := conversion.FormattingJsonSrc(lk.Video)
		info := recommendVideo{
			ID:            lk.ID,
			Uid:           lk.Uid,
			Title:         lk.Title,
			Video:         videoSrc,
			Cover:         cover,
			VideoDuration: lk.VideoDuration,
			Label:         conversion.StringConversionMap(lk.Label),
			Introduce:     lk.Introduce,
			Heat:          lk.Heat,
			BarrageNumber: len(lk.Barrage),
			Username:      lk.UserInfo.Username,
			CreatedAt:     lk.CreatedAt,
		}
		rl = append(rl, info)
	}
	response.RecommendList = rl
	return response
}

func GetVideoBarrageResponse(list *barrage.BarragesList) interface{} {
	barrageInfoList := make([][]interface{}, 0)
	for _, v := range *list {
		info := make([]interface{}, 0)
		info = append(info, v.Time)
		info = append(info, v.Type)
		info = append(info, v.Color)
		info = append(info, v.Author)
		info = append(info, v.Text)
		barrageInfoList = append(barrageInfoList, info)
	}
	return barrageInfoList
}

// 获取视频弹幕响应
type barrageInfo struct {
	Time     int       `json:"time"`
	Text     string    `json:"text"`
	SendTime time.Time `json:"sendTime"`
}

type barrageInfoList []barrageInfo

func GetVideoBarrageListResponse(list *barrage.BarragesList) interface{} {
	barrageList := make(barrageInfoList, 0)
	for _, v := range *list {
		info := barrageInfo{
			Time:     int(v.Time),
			Text:     v.Text,
			SendTime: v.PublicModel.CreatedAt,
		}
		barrageList = append(barrageList, info)
	}
	return barrageList
}

// 评论信息
type commentsInfo struct {
	ID               uint             `json:"id"`
	CommentID        uint             `json:"comment_id"`
	CommentIDContent string           `json:"comment_id_content"` //盖楼所回复的那条评论的内容
	CommentFirstID   uint             `json:"comment_first_id"`
	CreatedAt        time.Time        `json:"created_at"`
	Context          string           `json:"context"`
	Uid              uint             `json:"uid"`
	Username         string           `json:"username"`
	Photo            string           `json:"photo"`
	CommentUserID    uint             `json:"comment_user_id"`
	CommentUserName  string           `json:"comment_user_name"`
	LowerComments    commentsInfoList `json:"lowerComments"`
}

type commentsInfoList []*commentsInfo

type GetArticleContributionCommentsResponseStruct struct {
	Id             uint             `json:"id"`
	Comments       commentsInfoList `json:"comments"`
	CommentsNumber int              `json:"comments_number"`
}

// 得到分级结构
func (l commentsInfoList) getChildComment() commentsInfoList {
	topList := commentsInfoList{}
	for _, v := range l {
		if v.CommentID == 0 { //CommentID == 0说明是根评论
			topList = append(topList, v)
		}
	}
	//topListJson, _ := json.Marshal(topList)
	//global.Logger.Infof("所有根评论为%v", string(topListJson))
	return commentsInfoListSecondTree(topList, l)
}

// 生成树结构  params:
// menus：某个根评论的二级子评论；allData：全部评论
func commentsInfoListTree(menus commentsInfoList, allData commentsInfoList) commentsInfoList {
	//循环所有二级菜单
	for k, v := range menus {
		//查询所有该一级评论下的所有子评论
		var nodes commentsInfoList
		for _, av := range allData {
			if av.CommentID == v.ID {
				nodes = append(nodes, av)
			}
		}
		for _, node := range nodes { //将根评论的子评论
			menus[k].LowerComments = append(menus[k].LowerComments, node)
		}
		//LowerCommentsJson, _ := json.Marshal(menus[k].LowerComments)
		//global.Logger.Infof("ID为%v的一级子评论的子评论有：%v", menus[k].ID, string(LowerCommentsJson))
		//将刚刚查询出来的子菜单进行递归,查询出三级菜单和四级菜单
		commentsInfoListTree(nodes, allData)
	}
	return menus
}

// todo：修改这里的逻辑，使得子评论的子评论不再算作根评论的子评论，否则会重复计算
// menus:所有的一级评论；allData：全部评论信息
func commentsInfoListSecondTree(menus commentsInfoList, allData commentsInfoList) commentsInfoList {
	//循环所有一级菜单
	for k, v := range menus {
		//测试递归调用
		//res := handle(v.ID, allData)
		//global.Logger.Infoln("打印递归调用的返回结果")
		//for _, v := range res {
		//	global.Logger.Infof("res中的一个元素为%v", v)
		//}

		//var nodes commentsInfoList //二级评论
		//下面这是把直接的二级评论放在了二级评论层
		//todo:对于a每一个V.ID，应该递归去查找这个一级评论的所有子评论，全部放在menus[k].LowerComments中，这样就避免了多层树的结构
		//for _, av := range allData {
		//	if av.CommentID == v.ID {
		//		nodes = append(nodes, av)
		//	}
		//}

		var nodes commentsInfoList
		childComments := handle(v.ID, allData)
		for _, v := range childComments {
			nodes = append(nodes, v)
		}
		for _, node := range nodes {
			menus[k].LowerComments = append(menus[k].LowerComments, node)
		}

		//lowerCommentJson, _ := json.Marshal(menus[k].LowerComments)
		//global.Logger.Infof("ID为%v的根评论，其子评论有%v", menus[k].ID, string(lowerCommentJson))
		//
		//将刚刚查询出来的子菜单进行递归,查询出三级菜单和四级菜单
		//todo:下面这个暂时注释掉
		//commentsInfoListTree(nodes, allData)
	}
	return menus
}

// 递归查询一个根评论的所有子评论，先打印试试看对不对
func handle(rootId uint, allData commentsInfoList) commentsInfoList {
	var res commentsInfoList
	for _, av := range allData {
		if av.CommentID == rootId {
			res = append(res, av)
			list := handle(av.ID, allData)
			for _, v := range list {
				res = append(res, v)
			}
		}
	}

	return res
}

func GetVideoContributionCommentsResponse(vc *video.VideosContribution) GetArticleContributionCommentsResponseStruct {
	//评论
	comments := commentsInfoList{}
	for _, v := range vc.Comments {
		photo, _ := conversion.FormattingJsonSrc(v.UserInfo.Photo)
		commentUser := users.User{}
		commentUser.Find(v.CommentUserID)
		comments = append(comments, &commentsInfo{
			ID:              v.ID,
			CommentID:       v.CommentID,
			CommentFirstID:  v.CommentFirstID,
			CommentUserID:   v.CommentUserID,
			CommentUserName: commentUser.Username,
			CreatedAt:       v.CreatedAt,
			Context:         v.Context,
			Uid:             v.UserInfo.ID,
			Username:        v.UserInfo.Username,
			Photo:           photo,
		})
	}
	commentsList := comments.getChildComment()
	//输出
	response := GetArticleContributionCommentsResponseStruct{
		Id:             vc.ID,
		Comments:       commentsList,
		CommentsNumber: len(vc.Comments),
	}
	return response
}

type GetVideoManagementListItem struct {
	ID              uint      `json:"id"`
	Uid             uint      `json:"uid" `
	Title           string    `json:"title" `
	Video           string    `json:"video"`
	Cover           string    `json:"cover" `
	Reprinted       bool      `json:"reprinted"`
	CoverUrl        string    `json:"cover_url"`
	CoverUploadType string    `json:"cover_upload_type"`
	VideoDuration   int64     `json:"video_duration"`
	Label           []string  `json:"label"`
	Introduce       string    `json:"introduce"`
	Heat            int       `json:"heat"`
	BarrageNumber   int       `json:"barrageNumber"`
	CommentsNumber  int       `json:"comments_number"`
	CreatedAt       time.Time `json:"created_at"`
}

type GetVideoManagementList []GetVideoManagementListItem

func GetVideoManagementListResponse(vc *video.VideosContributionList) (interface{}, error) {
	list := make(GetVideoManagementList, 0)
	for _, v := range *vc {
		coverJson := new(common.Img)
		_ = json.Unmarshal(v.Cover, coverJson)
		cover, _ := conversion.FormattingJsonSrc(v.Cover)
		videoSrc, _ := conversion.FormattingJsonSrc(v.Video)
		info := GetVideoManagementListItem{
			ID:              v.ID,
			Uid:             v.Uid,
			Title:           v.Title,
			Video:           videoSrc,
			Cover:           cover,
			Reprinted:       conversion.Int8TurnBool(v.Reprinted),
			CoverUploadType: coverJson.Tp,
			CoverUrl:        coverJson.Src,
			VideoDuration:   v.VideoDuration,
			Label:           conversion.StringConversionMap(v.Label),
			Introduce:       v.Introduce,
			Heat:            v.Heat,
			BarrageNumber:   len(v.Barrage),
			CommentsNumber:  len(v.Comments),
			CreatedAt:       v.CreatedAt,
		}
		list = append(list, info)
	}
	return list, nil
}
