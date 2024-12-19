package video

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"haifengonline/consts"
	"haifengonline/global"
	"haifengonline/models/common"
	"haifengonline/models/contribution/video/barrage"
	"haifengonline/models/contribution/video/comments"
	"haifengonline/models/contribution/video/like"
	"haifengonline/models/users"
	"math/rand"
	"strconv"
	"time"
)

type VideosContribution struct {
	common.PublicModel
	Uid           uint           `json:"uid" gorm:"column:uid"`
	Title         string         `json:"title" gorm:"column:title"`
	Video         datatypes.JSON `json:"video" gorm:"column:video"` //默认1080p
	Video720p     datatypes.JSON `json:"video_720p" gorm:"column:video_720p"`
	Video480p     datatypes.JSON `json:"video_480p" gorm:"column:video_480p"`
	Video360p     datatypes.JSON `json:"video_360p" gorm:"column:video_360p"`
	MediaID       string         `json:"media_id" gorm:"column:media_id"`
	Cover         datatypes.JSON `json:"cover" gorm:"column:cover"`
	VideoDuration int64          `json:"video_duration" gorm:"column:video_duration"`
	Reprinted     int8           `json:"reprinted" gorm:"column:reprinted"`
	Label         string         `json:"label" gorm:"column:label"`
	Introduce     string         `json:"introduce" gorm:"column:introduce"`
	Heat          int            `json:"heat" gorm:"column:heat"`
	//todo:加了一个visible字段，可能会引起很多连锁反应
	IsVisible int `json:"is_visible" gorm:"column:is_visible"`

	UserInfo users.User           `json:"user_info" gorm:"foreignKey:Uid"`
	Likes    like.LikesList       `json:"likes" gorm:"foreignKey:VideoID" `
	Comments comments.CommentList `json:"comments" gorm:"foreignKey:VideoID"`
	Barrage  barrage.BarragesList `json:"barrage" gorm:"foreignKey:VideoID"`
}

type VideosContributionList []VideosContribution

func (VideosContribution) TableName() string {
	return "lv_video_contribution"
}

// Create 添加数据
func (vc *VideosContribution) Create() bool {
	err := global.Db.Create(&vc).Error
	if err != nil {
		return false
	}
	return true
}

// Delete 删除数据
func (vc *VideosContribution) Delete(id uint, uid uint) bool {
	//判断要删除的视频是否存在
	if global.Db.Where("id", id).Find(&vc).Error != nil {
		return false
	}
	//判断要删除的视频是否属于该用户
	if vc.Uid != uid {
		return false
	}
	if global.Db.Delete(&vc).Error != nil {
		return false
	}
	return true
}

// Update 更新数据
func (vc *VideosContribution) Update(info map[string]interface{}) bool {
	err := global.Db.Model(vc).Updates(info).Error
	if err != nil {
		return false
	}
	return true
}

func (vc *VideosContribution) Save() bool {
	err := global.Db.Save(vc).Error
	if err != nil {
		return false
	}
	return true
}

// FindByID 根据查询
func (vc *VideosContribution) FindByID(id uint) error {
	return global.Db.Where("id", id).Preload("Likes").Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Preload("UserInfo").Order("created_at desc")
	}).Preload("Barrage").Preload("UserInfo").Order("created_at desc").Find(&vc).Error
}

// GetVideoComments 获取评论
func (vc *VideosContribution) GetVideoComments(ID uint, info common.PageInfo) bool {
	err := global.Db.Where("id", ID).Preload("Likes").Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Preload("UserInfo").Order("created_at desc").Limit(info.Size).Offset((info.Page - 1) * info.Size)
	}).Find(vc).Error
	if err != nil {
		return false
	}
	return true
}

// Watch 添加播放
func (vc *VideosContribution) Watch(id uint) error {
	return global.Db.Debug().Model(vc).Where("id", id).Updates(map[string]interface{}{"heat": gorm.Expr("Heat  + ?", 1)}).Error
}

// GetVideoListBySpace 获取个人空间视频列表
func (vl *VideosContributionList) GetVideoListBySpace(id uint) error {
	return global.Db.Where("uid", id).Preload("Likes").Preload("Comments").Preload("Barrage").Order("created_at desc").Find(&vl).Error
}

// GetDiscussVideoCommentList 获取个人发布的视频和评论信息
func (vl *VideosContributionList) GetDiscussVideoCommentList(id uint) error {
	return global.Db.Where("uid", id).Preload("Comments").Find(&vl).Error
}

func (vl *VideosContributionList) GetVideoManagementList(info common.PageInfo, uid uint) error {
	return global.Db.Where("uid", uid).Preload("Likes").Preload("Comments").Preload("Barrage").Limit(info.Size).Offset((info.Page - 1) * info.Size).Order("created_at desc").Find(&vl).Error
}
func (vl *VideosContributionList) GetHoneVideoList(info common.PageInfo) error {
	//首页加载13个铺满后续15个
	var offset int
	if info.Page == 1 {
		info.Size = 11
		offset = (info.Page - 1) * info.Size
	}
	offset = (info.Page-2)*info.Size + 11

	//把这里修改为只返回is_visible=1的视频；然后设置延迟任务，在指定时间将视频可见设置为1
	//返回的视频流中：按照热度排序查询出info.size-5条数据，然后再随机抽取出5条数据，组合起来成为最终的返回结果
	var orderVideos []VideosContribution
	orderSize := info.Size - 5
	if orderSize > 0 {
		if err := global.Db.Preload("Likes").
			Preload("Comments").Preload("Barrage").
			Preload("UserInfo").Where("is_visible = ?", 1).
			Limit(info.Size).Offset(offset).Order("heat desc").Find(&orderVideos).Error; err != nil {
			return errors.New("failed to query videos  order by heat desc:" + err.Error())
		}
	}
	//将已经推荐的视频id存在bitmap中
	for _, video := range orderVideos {
		_, err := global.RedisDb.SetBit(fmt.Sprintf("%s%d", consts.UniqueVideoRecommendPrefix, -1), int64(video.ID), 1).Result()
		if err != nil {
			global.Logger.Errorf("set bitmap of orderVideos at home page  failed:%v", err)
		}
	}
	//随机查询10条数据，然后去重，最终只留下5条数据
	var randomVideo []VideosContribution
	if err := global.Db.Preload("Comments").Preload("Likes").Preload("Barrage").Preload("UserInfo").Where("is_visible = ?", 1).Order("RAND()").Limit(10).Find(&randomVideo).Error; err != nil {
		return errors.New("failed to query videos  randomly:" + err.Error())
	}

	//筛去已经推荐过的视频
	var appendRandomVideos []VideosContribution
	var skipVideos []VideosContribution
	for _, video := range randomVideo {
		result, err := global.RedisDb.GetBit(fmt.Sprintf("%s%d", consts.UniqueVideoRecommendPrefix, -1), int64(video.ID)).Result()
		if err != nil {
			global.Logger.Errorf("get bitmap failed:%v", err)
		}
		if result == 1 {
			skipVideos = append(skipVideos, video)
			continue
		} else {
			appendRandomVideos = append(appendRandomVideos, video)
		}
	}
	//长度不足5时，从已经筛掉的视频里随机抽取几个再填充到5为止
	if len(appendRandomVideos) < 5 {
		//从skipVideos中随机抽取几个填充进去
		for i := 0; i < 5-len(appendRandomVideos); i++ {
			appendRandomVideos = append(appendRandomVideos, skipVideos[rand.Intn(len(skipVideos))])
		}
	}

	*vl = append(orderVideos, appendRandomVideos...)

	return nil
	//return global.Db.Preload("Likes").Preload("Comments").Preload("Barrage").Preload("UserInfo").Where("is_visible = ?", 1).Limit(info.Size).Offset(offset).Order("heat desc").Find(&vl).Error
}

// GetRecommendList 获取推荐视频
func (vl *VideosContributionList) GetRecommendList(uid uint) error {
	//将每个用户对应的推荐列表都放在redis里缓存30s，避免不停的查询推荐视频
	result, err := global.RedisDb.Get(fmt.Sprintf("%s_%s", consts.RecommendVideosList, strconv.FormatUint(uint64(uid), 10))).Result()
	if err != nil && err != redis.Nil { //redis出错了，日志报告一下，然后继续查数据库
		global.Logger.Errorf("查询redis出错：%v", err)
	}
	if len(result) != 0 { //可以返回redis数据
		err := json.Unmarshal([]byte(result), vl)
		if err != nil {
			global.Logger.Errorf("类型转换出错:%v", err)
		}
		global.Logger.Infof("请求推荐视频数据，使用了redis缓存")
		return nil
	}

	//就是按照最新发布时间，筛选了7条视频出来
	//todo:获得用户专属的推荐视频，这个地方可以接入推荐系统，暂时还没想好该怎么做
	err = global.Db.Preload("Likes").Preload("Comments").Preload("Barrage").Preload("UserInfo").Order("created_at desc").Limit(7).Find(&vl).Error
	if err != nil {
		global.Logger.Errorf("查询数据库推荐视频出错：%v", err)
	}
	//写redis数据，并返回
	data, _ := json.Marshal(vl)
	global.RedisDb.Set(consts.RecommendVideosList, string(data), 30*time.Second)
	return nil
}

func (vl *VideosContributionList) Search(info common.PageInfo) error {
	return global.Db.Where("`title` LIKE ?", "%"+info.Keyword+"%").Preload("Likes").Preload("Comments").Preload("Barrage").Preload("UserInfo").Limit(info.Size).Offset((info.Page - 1) * info.Size).Order("created_at desc").Find(&vl).Error

}
