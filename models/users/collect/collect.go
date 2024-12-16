package collect

import (
	"haifengonline/global"
	"haifengonline/models/common"
	"haifengonline/models/contribution/video"
	"haifengonline/models/users"
)

// Collect 这是收藏夹与被收藏视频、所属用户的关联信息
type Collect struct {
	common.PublicModel
	Uid         uint `json:"uid" gorm:"column:uid"`                   //所属用户id
	FavoritesID uint `json:"favorites_id" gorm:"column:favorites_id"` //对应的收藏夹id
	VideoID     uint `json:"video_id" gorm:"column:video_id"`         //被收藏视频id

	UserInfo  users.User               `json:"userInfo" gorm:"foreignKey:Uid"`
	VideoInfo video.VideosContribution `json:"videoInfo" gorm:"foreignKey:VideoID"`
}

type CollectsList []Collect

func (Collect) TableName() string {
	return "lv_users_collect"
}

// Create 添加数据
func (c *Collect) Create() bool {
	err := global.Db.Debug().Create(c).Error
	if err != nil {
		return false
	}
	return true
}

// DetectByFavoritesID 删除收藏根据收藏夹id
func (c *Collect) DetectByFavoritesID(id uint) bool {
	err := global.Db.Where("favorites_id", id).Delete(c).Error
	if err != nil {
		return false
	}
	return true
}

func (l *CollectsList) FindVideoExistWhere(videoID uint) error {
	err := global.Db.Where("video_id", videoID).Find(l).Error
	return err
}

func (l *CollectsList) GetVideoInfo(id uint) error {
	err := global.Db.Where("favorites_id", id).Preload("VideoInfo").Find(l).Error
	return err
}

// FindIsCollectByFavorites 判断收藏夹列表内是否收藏该视频
func (l *CollectsList) FindIsCollectByFavorites(videoID uint, ids []uint) bool {
	//没创建收藏夹情况直接false
	if len(ids) == 0 {
		return false
	}
	err := global.Db.Where("video_id", videoID).Where("favorites_id", ids).Find(l).Error
	if err != nil {
		return false
	}
	if len(*l) == 0 {
		return false
	}
	return true
}
