package like

import (
	"haifengonline/models/common"
)

type Likes struct {
	common.PublicModel
	Uid       uint `json:"uid" gorm:"column:uid"`
	ArticleID uint `json:"article_id"  gorm:"column:article_id"`
}

type LikesList []Likes

func (Likes) TableName() string {
	return "lv_article_contribution_like"
}
