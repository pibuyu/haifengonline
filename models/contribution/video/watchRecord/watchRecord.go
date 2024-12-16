package watchRecord

import (
	"haifengonline/global"
)

type WatchRecord struct {
	Id           int64  `json:"id" gorm:"column:id"`
	Uid          uint   `json:"uid" gorm:"column:uid"`
	VideoID      uint   `json:"video_id"  gorm:"column:video_id"`
	WatchTime    string `json:"watch_time" gorm:"watch_time"`
	CreateTime   string `json:"create_time" gorm:"create_time"`
	DeleteStatus int    `json:"delete_status" gorm:"delete_status"`
}

func (WatchRecord) TableName() string {
	return "lv_watch_record"
}
func (r *WatchRecord) GetByUidAndVideoId(uid uint, videoId uint) error {
	return global.Db.Model(&WatchRecord{}).Where("uid = ? and video_id = ?", uid, videoId).Order("create_time desc").First(r).Error
}
