package contribution

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
	"haifengonline/consts"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/contribution/video"
	response "haifengonline/interaction/response/contribution/video"
	"haifengonline/logic/contribution/sokcet"
	"haifengonline/logic/users/notice"
	"haifengonline/models/common"
	"haifengonline/models/contribution/video"
	"haifengonline/models/contribution/video/barrage"
	"haifengonline/models/contribution/video/comments"
	"haifengonline/models/contribution/video/like"
	"haifengonline/models/contribution/video/watchRecord"
	transcodingTask "haifengonline/models/sundry/transcoding"
	"haifengonline/models/users/attention"
	"haifengonline/models/users/collect"
	"haifengonline/models/users/favorites"
	noticeModel "haifengonline/models/users/notice"
	"haifengonline/models/users/record"
	"haifengonline/utils/calculate"
	"haifengonline/utils/conversion"
	"haifengonline/utils/oss"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func SetIsVisibleById(id uint) error {
	global.Logger.Infof("当前时间为%s,将id为%d的视频设置为可见", time.Now().Format("2006-01-02 15:04:05"), id)
	if err := global.Db.Model(&video.VideosContribution{}).Where("id = ?", id).UpdateColumn("is_visible", 1).Error; err != nil {
		return err
	}
	return nil
}

// GetTop2HeatVideos 查询最热门的2条视频
func GetTop2HeatVideos() (results video.VideosContributionList, err error) {
	err = global.Db.Model(&video.VideosContribution{}).Order("heat DESC").Limit(2).Find(&results).Error
	if err != nil {
		global.Logger.Errorf("查询最热门的2条视频出错：%s", err.Error())
		return nil, errors.New("查询最热门的2条视频出错：" + err.Error())
	}
	global.Logger.Infoln("查询最热门的两个视频")
	return
}

// CreateVideoContribution 定时逻辑在上一层的调用函数里
func CreateVideoContribution(data *receive.CreateVideoContributionReceiveStruct, uid uint) (results interface{}, err error) {
	// 1.处理视频在数据库里的基本信息
	videoSrc, _ := json.Marshal(common.Img{
		Src: data.Video,
		Tp:  data.VideoUploadType,
	})
	coverImg, _ := json.Marshal(common.Img{
		Src: data.Cover,
		Tp:  data.CoverUploadType,
	})
	var width, height int
	// Deprecated:本地上传基本弃用了,都上传到oss里
	if data.VideoUploadType == "local" {
		//如果是本地上传
		width, height, err = calculate.GetVideoResolution(data.Video)
		if err != nil {
			global.Logger.Error("获取视频分辨率失败")
			return nil, fmt.Errorf("获取视频分辨率失败")
		}
	} else {
		mediaInfo, err := oss.GetMediaInfo(data.Media)
		if err != nil {
			return nil, errors.New("获取视频信息失败，错误：" + err.Error())
		}
		width, _ = strconv.Atoi(*mediaInfo.Body.MediaInfo.FileInfoList[0].FileBasicInfo.Width)
		height, _ = strconv.Atoi(*mediaInfo.Body.MediaInfo.FileInfoList[0].FileBasicInfo.Height)
	}
	videoContribution := &video.VideosContribution{
		Uid:           uid,
		Title:         data.Title,
		Cover:         coverImg,
		Reprinted:     conversion.BoolTurnInt8(*data.Reprinted),
		Label:         conversion.MapConversionString(data.Label),
		VideoDuration: data.VideoDuration,
		Introduce:     data.Introduce,
		Heat:          0,
	}
	if data.Media != nil {
		videoContribution.MediaID = *data.Media
	}
	// 2.定义分辨率列表，在这里先将每个分辨率的视频源都设置为初始源
	resolutions := []int{1080, 720, 480, 360}
	if height >= 1080 {
		resolutions = resolutions[1:]
		videoContribution.Video = videoSrc
	} else if height >= 720 {
		resolutions = resolutions[2:]
		videoContribution.Video720p = videoSrc
	} else if height >= 480 {
		resolutions = resolutions[3:]
		videoContribution.Video480p = videoSrc
	} else if height >= 360 {
		resolutions = resolutions[4:]
		videoContribution.Video360p = videoSrc
	} else {
		global.Logger.Error("上传视频分辨率过低")
		return nil, fmt.Errorf("上传视频分辨率过低")
	}

	//todo:当没有传递定时创建的时间时，默认is_visible=1
	if data.DateTime == "" {
		videoContribution.IsVisible = 1
	}
	if !videoContribution.Create() {
		return nil, fmt.Errorf("保存失败")
	}

	// 3.进行视频转码
	go func(width, height int, video *video.VideosContribution) {
		// 3.1 本地上传的视频，使用 ffmpeg 处理
		if data.VideoUploadType == "local" {
			inputFile := data.Video
			sr := strings.Split(inputFile, ".")
			//对每个清晰度选项，将视频转为对应的清晰度并保存src
			for _, r := range resolutions {
				// 计算转码后的宽和高需要取整
				w := int(math.Ceil(float64(r) / float64(height) * float64(width)))
				h := int(math.Ceil(float64(r)))
				if h >= height {
					continue
				}
				dst := sr[0] + fmt.Sprintf("_output_%dp."+sr[1], r)
				cmd := exec.Command("ffmpeg",
					"-i",
					inputFile,
					"-vf",
					fmt.Sprintf("scale=%d:%d", w, h),
					"-c:a",
					"copy",
					"-c:v",
					"libx264",
					"-preset",
					"medium",
					"-crf",
					"23",
					"-y",
					dst)
				err = cmd.Run()
				if err != nil {
					global.Logger.Errorf("视频: %s :转码 %d*%d 失败。command : %s ,err info :%s", inputFile, w, h, cmd, err)
					continue
				}
				src, _ := json.Marshal(common.Img{
					Src: dst,
					Tp:  "local",
				})
				switch r {
				case 1080:
					videoContribution.Video = src
				case 720:
					videoContribution.Video720p = src
				case 480:
					videoContribution.Video480p = src
				case 360:
					videoContribution.Video360p = src
				}
				if !videoContribution.Save() {
					global.Logger.Errorf("视频 :%s : 转码%d*%d后视频保存到数据库失败", inputFile, w, h)
				}
				global.Logger.Infof("视频 :%s : 转码%d*%d成功", inputFile, w, h)
			}
		} else if data.VideoUploadType == "aliyunOss" && global.Config.AliyunOss.IsOpenTranscoding {
			// 3.2 上传到阿里云的视频，调用iceClient提供的接口进行处理，并调用AliyunTranscodingMedia回调函数进行处理
			inputFile := data.Video
			sr := strings.Split(inputFile, ".")
			//云转码处理
			for _, r := range resolutions {
				//获取转码模板
				var template string
				dst := sr[0] + fmt.Sprintf("_output_%dp."+sr[1], r)
				src, _ := json.Marshal(common.Img{
					Src: dst,
					Tp:  data.VideoUploadType,
				})
				switch r {
				case 1080:
					template = global.Config.AliyunOss.TranscodingTemplate1080p
					videoContribution.Video = src
				case 720:
					template = global.Config.AliyunOss.TranscodingTemplate720p
					videoContribution.Video720p = src
				case 480:
					template = global.Config.AliyunOss.TranscodingTemplate480p
					videoContribution.Video480p = src
				case 360:
					template = global.Config.AliyunOss.TranscodingTemplate360p
					videoContribution.Video360p = src
				}
				outputUrl, _ := conversion.SwitchIngStorageFun(data.VideoUploadType, dst)
				taskName := "转码 : " + *data.Media + "时间 :" + time.Now().Format("2006.01.02 15:04:05") + " template : " + template
				jobInfo, err := oss.SubmitTranscodeJob(taskName, video.MediaID, outputUrl, template)
				if err != nil {
					global.Logger.Errorf("视频云转码 : %s 失败 err : %s", outputUrl, err.Error())
					continue
				}
				task := &transcodingTask.TranscodingTask{
					TaskID:     *jobInfo.TranscodeParentJob.ParentJobId,
					VideoID:    video.ID,
					Resolution: r,
					Dst:        dst,
					Status:     0,
					Type:       transcodingTask.Aliyun,
				}
				if !task.AddTask() {
					global.Logger.Errorf("视频云转码任务名: %s 后将视频任务 保存到数据库失败", taskName)
				}
			}
		}
	}(width, height, videoContribution)

	return videoContribution.ID, nil
}

func UpdateVideoContribution(data *receive.UpdateVideoContributionReceiveStruct, uid uint) (results interface{}, err error) {
	//更新视频
	videoInfo := new(video.VideosContribution)
	err = videoInfo.FindByID(data.ID)
	if err != nil {
		return nil, fmt.Errorf("操作视频不存在")
	}
	// 判断这个视频是不是这个用户发布的
	if videoInfo.Uid != uid {
		return nil, fmt.Errorf("非法操作")
	}
	// 将封面img信息转为json串,存在数据库里，需要用的时候再UnMarshal转换为结构体
	coverImg, _ := json.Marshal(common.Img{
		Src: data.Cover,
		Tp:  data.CoverUploadType,
	})
	updateList := map[string]interface{}{
		"cover":     coverImg,
		"title":     data.Title,
		"label":     conversion.MapConversionString(data.Label),
		"reprinted": conversion.BoolTurnInt8(*data.Reprinted),
		"introduce": data.Introduce,
	}
	//进行视频资料更新
	if !videoInfo.Update(updateList) {
		return nil, fmt.Errorf("更新数据失败")
	}
	return "更新成功", nil
}

func DeleteVideoByPath(data *receive.DeleteVideoByPathReceiveStruct) (results interface{}, err error) {
	err = oss.DeleteOSSFile([]string{data.Path})
	if err != nil {
		return nil, fmt.Errorf("删除oss对象失败：%v", err)
	}
	return "删除成功", nil
}

func GetLastWatchTime(uid uint, videoId uint) (result interface{}, err error) {
	//在数据库里查到观看进度
	record := &watchRecord.WatchRecord{}
	if err := record.GetByUidAndVideoId(uid, videoId); err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询观看进度failed:%v", err)
	}
	return record.WatchTime, nil
}

func SendWatchTime(data *receive.SendWatchTimeReqStruct, uid uint) error {
	videoId, _ := strconv.ParseUint(data.Id, 10, 64)
	var record watchRecord.WatchRecord
	global.Db.Model(&watchRecord.WatchRecord{}).Where("uid = ? and video_id = ?", uid, data.Id).First(&record)
	if record.Id != 0 {
		record.WatchTime = data.Time
		return global.Db.Model(&watchRecord.WatchRecord{}).Where("uid = ? and video_id = ?", uid, data.Id).Updates(record).Error
	}

	record.Uid = uid
	record.VideoID = uint(videoId)
	record.WatchTime = data.Time
	record.CreateTime = time.Now().Format("2006-01-02 15:04:05")

	return global.Db.Model(&watchRecord.WatchRecord{}).Save(&record).Error
}

func DeleteVideoByID(data *receive.DeleteVideoByIDReceiveStruct, uid uint) (results interface{}, err error) {
	vo := new(video.VideosContribution)

	//先查到对应视频在oss里的位置
	err = vo.FindByID(data.ID)
	if err != nil {
		global.Logger.Errorf("查找视频在oss中的路径失败：%v", err)
	}
	var deleteVideoPaths []string
	if vo.Video720p != nil {
		//vo.Video720p.String()
		deleteVideoPaths = append(deleteVideoPaths, vo.Video720p.String())
	} else if vo.Video != nil {
		deleteVideoPaths = append(deleteVideoPaths, vo.Video.String())
	} else if vo.Video480p != nil {
		deleteVideoPaths = append(deleteVideoPaths, vo.Video480p.String())
	} else if vo.Video360p != nil {
		deleteVideoPaths = append(deleteVideoPaths, vo.Video360p.String())
	} else {
		return nil, fmt.Errorf("视频不存在")
	}

	if !vo.Delete(data.ID, uid) {
		return nil, fmt.Errorf("删除失败")
	}
	//去oss里删除对应的视频
	if err = oss.DeleteOSSFile(deleteVideoPaths); err != nil {
		return nil, fmt.Errorf("删除oss中的视频失败:%v", err)
	}
	//global.Logger.Infof("用户 %d 在oss中删除了视频：%s", uid, strings.Join(deleteVideoPaths, ","))
	return "删除成功", nil
}

func GetVideoContributionByID(data *receive.GetVideoContributionByIDReceiveStruct, uid uint) (results interface{}, err error) {
	//观看视频的同时将此视频id放进bitmap中，推荐视频的时候随机请求，然后滤除掉最近推荐过的和观看过的
	key := fmt.Sprintf("%s%d", consts.UniqueVideoRecommendPrefix, -1)
	_, err = global.RedisDb.SetBit(key, int64(data.VideoID), 1).Result()
	if err != nil {
		global.Logger.Errorf("set bitmap failed:%v", err)
	}
	videoInfo := new(video.VideosContribution)
	//获取视频信息
	err = videoInfo.FindByID(data.VideoID)
	if err != nil {
		return nil, fmt.Errorf("查询信息失败")
	}
	isAttention := false
	isLike := false
	isCollect := false
	if uid != 0 { //带有用户信息的请求才能加播放量
		//进行视频播放增加:先去redis添加这个videoWatchBy_videoId : uid的键值对，然后去给数据库中的播放量+1
		//todo:这里有个bug：当redis缓存里有这个视频的信息时，点击视频不增加播放量    bug的原因是加载播放页面时useInit函数和vidoe/video.vue的onMounted各调用了一次GetVideoContributionByID方法，就导致heat递增2次
		if !global.RedisDb.SIsMember(consts.VideoWatchByID+strconv.Itoa(int(data.VideoID)), uid).Val() { //SIsMember key value :查询redis中是否存在这个键值对
			//最近无播放
			global.RedisDb.SAdd(consts.VideoWatchByID+strconv.Itoa(int(data.VideoID)), uid)
			if videoInfo.Watch(data.VideoID) != nil { //这里更新的是数据库
				global.Logger.Error("添加播放量错误视频video_id:", data.VideoID)
			}
			videoInfo.Heat++ //这里更新的是已经拿到的对象
		}
		//获取是否关注
		at := new(attention.Attention)
		isAttention = at.IsAttention(uid, videoInfo.UserInfo.ID)

		//获取是否关注
		lk := new(like.Likes)
		isLike = lk.IsLike(uid, videoInfo.ID)

		//判断是否已经收藏
		fl := new(favorites.FavoriteList) //FavoriteList是收藏夹
		err = fl.GetFavoritesList(uid)
		if err != nil {
			return nil, fmt.Errorf("查询失败")
		}
		flIDs := make([]uint, 0)
		for _, v := range *fl {
			flIDs = append(flIDs, v.ID) //v.ID是收藏夹的ID
		}
		//判断是否在收藏夹内
		cl := new(collect.CollectsList)
		isCollect = cl.FindIsCollectByFavorites(data.VideoID, flIDs) //

		//添加历史记录
		rd := new(record.Record)
		err = rd.AddVideoRecord(uid, data.VideoID)
		if err != nil {
			return nil, fmt.Errorf("添加历史记录失败")
		}

	}
	//获取推荐列表
	recommendList := new(video.VideosContributionList)
	err = recommendList.GetRecommendList(uid)
	if err != nil {
		return nil, err
	}
	res := response.GetVideoContributionByIDResponse(videoInfo, recommendList, isAttention, isLike, isCollect)
	return res, nil
}

func SendVideoBarrage(data *receive.SendVideoBarrageReceiveStruct, uid uint) (results interface{}, err error) {
	//获取弹幕list的时候先查了缓存，根据cache aside，这里应该先修改数据库后删除缓存
	if global.Filter.IsSensitive(data.Text) {
		return nil, fmt.Errorf("您输入的弹幕包含敏感词")
	}
	//保存弹幕
	videoID, _ := strconv.ParseUint(data.ID, 0, 19)
	bg := barrage.Barrage{
		Uid:     uid,
		VideoID: uint(videoID),
		Time:    data.Time,
		Author:  data.Author,
		Type:    data.Type,
		Text:    data.Text,
		Color:   data.Color,
	}
	if !bg.Create() {
		return data, fmt.Errorf("发送弹幕失败")
	}
	//删除缓存
	key := fmt.Sprintf("%s%s", consts.VideoBarragePrefix, data.ID)
	_ = global.RedisDb.Del(key)
	//socket消息通知
	res := sokcet.ChanInfo{
		Type: consts.VideoSocketTypeResponseBarrageNum,
		Data: nil,
	}
	for _, v := range sokcet.Severe.VideoRoom[uint(videoID)] {
		v.MsgList <- res
	}

	return data, nil
}

func GetVideoBarrage(data *receive.GetVideoBarrageReceiveStruct) (results interface{}, err error) {
	//获取视频弹幕
	list := new(barrage.BarragesList)

	videoID, _ := strconv.ParseUint(data.ID, 0, 19)
	if !list.GetVideoBarrageByID(uint(videoID)) {
		return nil, fmt.Errorf("查询失败")
	}

	res := response.GetVideoBarrageResponse(list)
	return res, nil
}

func GetVideoBarrageList(data *receive.GetVideoBarrageListReceiveStruct) (results interface{}, err error) {
	//获取视频弹幕
	list := &barrage.BarragesList{}

	//原本的查redis缓存
	key := fmt.Sprintf("%s%s", consts.VideoBarragePrefix, data.ID)
	result, err := global.RedisDb.Get(key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		global.Logger.Errorf("获取视频%v弹幕信息时，查询redis err:%v", data.ID, err)
	}
	if len(result) != 0 { //如果查到了缓存，直接封装返回即可
		if err := json.Unmarshal([]byte(result), list); err == nil {
			global.Logger.Infof("查询视频%s弹幕信息时，命中cache返回：%v", data.ID, list)
			return list, nil
		}
		global.Logger.Errorf("获取视频%v弹幕信息时，解封装错误", data.ID)
	}

	videoID, _ := strconv.ParseUint(data.ID, 0, 19)
	if !list.GetVideoBarrageByID(uint(videoID)) {
		return nil, fmt.Errorf("查询失败")
	}

	//set redis
	bytes, err := json.Marshal(list)
	if err == nil {
		global.RedisDb.Set(key, bytes, 5*time.Second)
	} else {
		global.Logger.Errorf("封装视频%v弹幕信息错误:%v", data.ID, err)
	}

	res := response.GetVideoBarrageListResponse(list)
	return res, nil
}

func VideoPostComment(data *receive.VideosPostCommentReceiveStruct, uid uint) (results interface{}, err error) {
	//todo:在这里进行敏感内容的过滤
	if global.Filter.IsSensitive(data.Content) {
		return nil, fmt.Errorf("评论内容含有敏感词")
	}
	videoInfo := new(video.VideosContribution)
	err = videoInfo.FindByID(data.VideoID)
	if err != nil {
		return nil, fmt.Errorf("视频不存在")
	}
	//找到被评论的那条评论
	ct := comments.Comment{
		PublicModel: common.PublicModel{ID: data.ContentID},
	}
	//获取被评论的那条评论所属的根评论
	CommentFirstID := ct.GetCommentFirstID()

	ctu := comments.Comment{
		PublicModel: common.PublicModel{ID: data.ContentID},
	}
	CommentUserID := ctu.GetCommentUserID()
	comment := comments.Comment{
		Uid:            uid,
		VideoID:        data.VideoID,
		Context:        data.Content,
		CommentID:      data.ContentID,
		CommentUserID:  CommentUserID,
		CommentFirstID: CommentFirstID,
	}
	if !comment.Create() {
		return nil, fmt.Errorf("发布失败")
	}

	//socket推送(在线的情况下)
	if _, ok := notice.Severe.UserMapChannel[videoInfo.UserInfo.ID]; ok {
		userChannel := notice.Severe.UserMapChannel[videoInfo.UserInfo.ID]
		userChannel.NoticeMessage(noticeModel.VideoComment)
	}

	return "发布成功", nil
}

func GetVideoComment(data *receive.GetVideoCommentReceiveStruct) (results interface{}, err error) {
	videosContribution := new(video.VideosContribution)
	if !videosContribution.GetVideoComments(data.VideoID, data.PageInfo) {
		return nil, fmt.Errorf("查询失败")
	}
	return response.GetVideoContributionCommentsResponse(videosContribution), nil
}

func GetVideoManagementList(data *receive.GetVideoManagementListReceiveStruct, uid uint) (results interface{}, err error) {
	//获取个人发布视频信息
	list := new(video.VideosContributionList)
	err = list.GetVideoManagementList(data.PageInfo, uid)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	res, err := response.GetVideoManagementListResponse(list)
	if err != nil {
		return nil, fmt.Errorf("响应失败")
	}
	return res, nil
}

// todo:重构点赞模块
// 最重要的是用户对哪些稿件进行点赞的点赞记录表和一个稿件实体被点过多少赞和踩的点赞计数表
func LikeVideo(data *receive.LikeVideoReceiveStruct, uid uint) (results interface{}, err error) {
	//点赞视频
	//参数：video_id和uid
	videoInfo := new(video.VideosContribution)
	err = videoInfo.FindByID(data.VideoID)
	if err != nil {
		return nil, fmt.Errorf("视频不存在")
	}
	lk := new(like.Likes)
	err = lk.Like(uid, data.VideoID, videoInfo.UserInfo.ID)
	if err != nil {
		return nil, fmt.Errorf("操作失败")
	}

	//socket推送(在线的情况下)
	//当userMapChannel能查到这个作者时，也就是用户在线的时候
	//todo:这里也可以修改为异步推送点赞消息;然后还需要解决加入点赞者和视频作者不在同一台服务器的socket连接上该怎么处理
	if _, ok := notice.Severe.UserMapChannel[videoInfo.UserInfo.ID]; ok {
		userChannel := notice.Severe.UserMapChannel[videoInfo.UserInfo.ID]
		//传递过去的参数是"videoLike"字符串
		userChannel.NoticeMessage(noticeModel.VideoLike)
	}

	return "操作成功", nil
}

func LikeVideoComment(data *receive.LikeVideoCommentReqStruct) (results interface{}, err error) {
	//todo:在这里实现对点赞的聚合写入，以及更新redis中评论的点赞数
	//先不管怎么获取到评论，直接更新redis并实现聚合写入先
	zsetKey := fmt.Sprintf("%s%s", consts.VideoCommentZSetPrefix, strconv.Itoa(data.VideoCommentId))
	hashKey := fmt.Sprintf("%s%s", consts.VideoCommentHashPrefix, strconv.Itoa(data.VideoCommentId))

	//判断是否在zset里应该用整条评论内容做member值，因为放进去的时候就是这样放的
	_, err = global.RedisDb.HGet(hashKey, strconv.Itoa(data.VideoCommentId)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		global.Logger.Errorf("在hash里查询评论点赞数failed:%v", err)
	}
	if err == redis.Nil { //说明这条评论没有放进hash中去
		global.Logger.Infof("向hash和zset里添加id=%d的评论", data.VideoCommentId)
		com := &comments.Comment{}
		if err := global.Db.Model(&comments.Comment{}).Where("id = ?", data.VideoCommentId).Find(&com).Error; err != nil {
			global.Logger.Errorf("从数据库查询id=%d的评论内容失败: %v", data.VideoCommentId, err)
		}
		com.Heat++
		_, err := global.RedisDb.HSet(hashKey, strconv.Itoa(data.VideoCommentId), com).Result()
		if err != nil {
			global.Logger.Errorf("向hash中放入id=%d的评论内容失败: %v", data.VideoCommentId, err)
		}

		jsonComment, _ := json.Marshal(com)
		_, err = global.RedisDb.ZAdd(zsetKey, redis.Z{
			Member: jsonComment,
			Score:  float64(com.Heat),
		}).Result()
		if err != nil {
			global.Logger.Errorf("在zset里增加评论 %v failed:%v", com, err)
		}
	} else {
		global.Logger.Infof("更新id=%d的评论", data.VideoCommentId)
		//根据commentId从hash中拿到评论内容，然后将其点赞数+1
		result, err := global.RedisDb.HGet(hashKey, strconv.Itoa(data.VideoCommentId)).Result()
		if err != nil {
			global.Logger.Errorf("从hash中获取id=%d的评论failed:%v", data.VideoCommentId, err)
		}
		_, err = global.RedisDb.ZIncrBy(zsetKey, 1, result).Result()
		if err != nil {
			global.Logger.Errorf("向zset中评论 ： %v 的热度+1failed: %v", result, err)
		}

	}

	//_, err = global.RedisDb.ZScore(zsetKey, strconv.Itoa(data.VideoCommentId)).Result()
	//if err != nil && !errors.Is(err, redis.Nil) {
	//	global.Logger.Errorf("在zset里查询评论点赞数failed:%v", err)
	//}
	//
	//if err == redis.Nil {
	//	//zset里不存在这条评论,查到这条评论的信息然后放入zset
	//	com := &comments.Comment{}
	//	com.Find(uint(data.VideoCommentId))
	//	com.Heat++
	//	//转为json存储
	//	commentJson, err := json.Marshal(com)
	//	if err != nil {
	//		global.Logger.Errorf("将评论 %d 转为json失败:%v", data.VideoCommentId, err)
	//	}
	//	//添加到zset里
	//	_, err = global.RedisDb.ZAdd(zsetKey, redis.Z{
	//		Score:  float64(com.Heat),
	//		Member: commentJson, //member是一整条的评论内容
	//	}).Result()
	//	global.Logger.Infof("将id=%d的评论放入zset中去", data.VideoCommentId)
	//	if err != nil {
	//		global.Logger.Errorf("在zset里增加评论 %v 数failed:%v", com, err)
	//	}
	//
	//} else {
	//	//存在这条评论，直接将被赞数+1
	//	_, err := global.RedisDb.ZIncrBy(zsetKey, 1, strconv.Itoa(data.VideoCommentId)).Result()
	//	if err != nil {
	//		global.Logger.Errorf("增加zset里id= %d 的评论被赞数failed: %v", data.VideoCommentId, err)
	//	}
	//	global.Logger.Infof("将zset中id=%d的评论热度+1", data.VideoCommentId)
	//}
	return 1, nil
}
