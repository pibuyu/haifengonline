package comments

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"haifengonline/global"
	"haifengonline/models/common"
	"haifengonline/models/users"
	"haifengonline/models/users/notice"
)

type Comment struct {
	common.PublicModel
	Uid            uint   `json:"uid" gorm:"column:uid"`
	VideoID        uint   `json:"video_id" gorm:"column:video_id"`
	Context        string `json:"context" gorm:"column:context"`
	CommentID      uint   `json:"comment_id" gorm:"column:comment_id"`
	CommentUserID  uint   `json:"comment_user_id" gorm:"column:comment_user_id"`
	CommentFirstID uint   `json:"comment_first_id" gorm:"column:comment_first_id"`

	UserInfo  users.User `json:"user_info" gorm:"foreignKey:Uid"`
	VideoInfo VideoInfo  `json:"video_info" gorm:"foreignKey:VideoID"`
}
type CommentList []Comment

func (Comment) TableName() string {
	return "lv_video_contribution_comments"
}

// VideoInfo 临时加一个video模型解决依赖循环
type VideoInfo struct {
	common.PublicModel
	Uid   uint           `json:"uid" gorm:"uid"`
	Title string         `json:"title" gorm:"title"`
	Video datatypes.JSON `json:"video" gorm:"video"`
	Cover datatypes.JSON `json:"cover" gorm:"cover"`
}

func (VideoInfo) TableName() string {
	return "lv_video_contribution"
}

// Find 根据id 查询
func (c *Comment) Find(id uint) {
	_ = global.Db.Where("id", id).Find(&c).Error
}

// Create 添加数据
func (c *Comment) Create() bool {
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		videoInfo := new(VideoInfo)
		err := tx.Where("id", c.VideoID).Find(videoInfo).Error
		if err != nil {
			return err
		}
		err = tx.Create(&c).Error
		if err != nil {
			return err
		}
		//消息通知
		if videoInfo.Uid == c.Uid {
			return nil
		}
		//添加消息通知
		ne := new(notice.Notice)
		err = ne.AddNotice(videoInfo.Uid, c.Uid, videoInfo.ID, notice.VideoComment, c.Context)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return false
	}
	return true
}

// GetCommentFirstID 获取最顶层的评论(子评论所属的根评论)id
func (c *Comment) GetCommentFirstID() uint {
	_ = global.Db.Where("id", c.ID).Find(&c).Error
	if c.CommentID != 0 {
		c.ID = c.CommentID
		c.GetCommentFirstID()
	}
	return c.ID
}

// GetCommentUserID 获取评论id的user
func (c *Comment) GetCommentUserID() uint {
	_ = global.Db.Where("id", c.ID).Find(&c).Error
	return c.Uid
}

func (cl *CommentList) GetCommentListByIDs(ids []uint, info common.PageInfo) error {
	return global.Db.Where("video_id", ids).Preload("UserInfo").Preload("VideoInfo").Offset((info.Page - 1) * info.Size).Limit(info.Size).Order("created_at desc").Find(&cl).Error
}
