package like

import (
	"gorm.io/gorm"
	"haifengonline/global"
	"haifengonline/models/common"
	"haifengonline/models/users/notice"
)

type Likes struct {
	common.PublicModel
	Uid     uint `json:"uid" gorm:"column:uid"`
	VideoID uint `json:"video_id"  gorm:"column:video_id"`
}

type LikesList []Likes

func (Likes) TableName() string {
	return "lv_video_contribution_like"
}

func (l *Likes) IsLike(uid uint, videoID uint) bool {
	err := global.Db.Where(Likes{Uid: uid, VideoID: videoID}).Find(l).Error
	if err != nil {
		return false
	}
	if l.ID <= 0 {
		return false
	}
	return true
}
func (l *Likes) Like(uid uint, videoID uint, videoUid uint) error {
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("uid", uid).Where("video_id", videoID).Find(l).Error
		if err != nil {
			return err
		}
		if l.ID > 0 { //以及点过赞，删除消息通知(点赞的时候往notice表插入一条type=videoLike的消息，这是取消点赞，自然要删除)
			err = tx.Where("uid", uid).Where("video_id", videoID).Delete(l).Error
			if err != nil {
				return err
			}
			//点赞自己作品不进行通知
			if videoUid == uid {
				return nil
			}

			ne := new(notice.Notice)
			err = ne.Delete(videoUid, uid, videoID, notice.VideoLike)
			if err != nil {
				return err
			}
		} else { //没有点过赞，就在like表里插入一条记录；同时往notice表插入一条type=videoLike的消息
			l.Uid = uid
			l.VideoID = videoID
			err = global.Db.Create(l).Error
			if err != nil {
				return err
			}
			//点赞自己作品不进行通知
			if videoUid == uid {
				return nil
			}
			//添加消息通知,点赞的时候往notice表插入一条type=videoLike的消息
			ne := new(notice.Notice)
			err = ne.AddNotice(videoUid, uid, videoID, notice.VideoLike, "赞了您的作品")
			if err != nil {
				return err
			}
		}
		// 返回 nil 提交事务
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
