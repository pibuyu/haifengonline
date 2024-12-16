package favorites

import (
	"fmt"
	"gorm.io/datatypes"
	"haifengonline/global"
	"haifengonline/models/common"
	"haifengonline/models/users"
	"haifengonline/models/users/collect"
)

// Favorites 这是收藏夹的基础信息
type Favorites struct {
	common.PublicModel
	Uid     uint           `json:"uid" gorm:"column:uid"`                //所属用户id
	Title   string         `json:"title" gorm:"column:title"`            //收藏夹名称
	Content string         `json:"content" gorm:"column:content"`        //收藏夹简介
	Cover   datatypes.JSON `json:"cover" gorm:"type:json;comment:cover"` //收藏夹封面图片链接
	Max     int            `json:"max" gorm:"column:max"`                //单个收藏夹最大收藏视频数

	UserInfo    users.User           `json:"userInfo" gorm:"foreignKey:Uid"`
	CollectList collect.CollectsList `json:"collectList"  gorm:"foreignKey:FavoritesID"`
}

type FavoriteList []Favorites

func (Favorites) TableName() string {
	return "lv_users_favorites"
}

// Find 查询
func (f *Favorites) Find(id uint) bool {
	err := global.Db.Where("id", id).Preload("CollectList").Order("created_at desc").Find(&f).Error
	if err != nil {
		return false
	}
	return true
}

// Create 添加数据
func (f *Favorites) Create() bool {
	err := global.Db.Create(&f).Error
	if err != nil {
		return false
	}
	return true
}

// AloneTitleCreate 单标题创建
func (f *Favorites) AloneTitleCreate() bool {
	err := global.Db.Create(&f).Error
	if err != nil {
		return false
	}
	return true
}

// Update 更新数据
func (f *Favorites) Update() bool {
	err := global.Db.Updates(&f).Error
	if err != nil {
		return false
	}
	return true
}

// Delete 删除数据
func (f *Favorites) Delete(id uint, uid uint) error {
	err := global.Db.Where("id", id).Find(&f).Error
	if err != nil {
		return fmt.Errorf("查询失败")
	}
	if f.ID <= 0 {
		return fmt.Errorf("收藏夹不存在")
	}
	err = global.Db.Delete(&f).Error
	if f.Uid != uid {
		return fmt.Errorf("非创建者不可删除")
	}
	//删除收藏记录
	cl := new(collect.Collect)
	if !cl.DetectByFavoritesID(id) {
		return fmt.Errorf("删除收藏记录失败")
	}
	if err != nil {
		return fmt.Errorf("删除失败")
	}
	return nil
}

func (fl *FavoriteList) GetFavoritesList(id uint) error {
	//这里两个Preload字段的作用：因为FavoriteList结构体中有UserInfo和CollectList结构，因此关联这两张表，将查询到的UserInfo和CollectList信息填充到FavoriteList结构的对应字段中
	err := global.Db.Where("uid", id).Preload("UserInfo").Preload("CollectList").Order("created_at desc").Find(fl).Error
	if err != nil {
		return err
	}
	return nil
}
