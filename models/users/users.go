package users

import (
	"crypto/md5"
	"haifengonline/global"
	"haifengonline/models/common"
	"haifengonline/models/users/liveInfo"

	"fmt"
	"gorm.io/datatypes"
	"time"
)

// User 表结构体
type User struct {
	common.PublicModel
	Email       string         `json:"email" gorm:"column:email"`
	Username    string         `json:"username" gorm:"column:username"`
	Openid      string         `json:"openid" gorm:"column:openid"`
	Salt        string         `json:"salt" gorm:"column:salt"`
	Password    string         `json:"password" gorm:"column:password"`
	Photo       datatypes.JSON `json:"photo" gorm:"column:photo"`
	Gender      int8           `json:"gender" gorm:"column:gender"`
	BirthDate   time.Time      `json:"birth_date" gorm:"column:birth_date"`
	IsVisible   int8           `json:"is_visible" gorm:"column:is_visible"`
	Signature   string         `json:"signature" gorm:"column:signature"`
	SocialMedia string         `json:"social_media" gorm:"social_media"`

	LiveInfo liveInfo.LiveInfo `json:"liveInfo" gorm:"foreignKey:Uid"`
}

type UserList []User

func (User) TableName() string {
	return "lv_users"
}

// Update 更新数据
func (us *User) Update() bool {
	err := global.Db.Where("id", us.ID).Updates(&us).Error
	if err != nil {
		return false
	}
	return true
}

// UpdatePureZero 更新数据存在0值
func (us *User) UpdatePureZero(user map[string]interface{}) bool {
	global.Logger.Infof("用户%d修改了自己的信息->name:%s,gender:%d,birthdate:%s,social_media:%s,signature:%s", us.ID, us.Username, us.Gender, us.BirthDate, us.SocialMedia, us.Signature)
	err := global.Db.Model(&us).Where("id", us.ID).Updates(user).Error
	if err != nil {
		return false
	}
	return true
}

// Create 添加数据
func (us *User) Create() bool {
	err := global.Db.Create(&us).Error
	if err != nil {
		return false
	}
	return true
}

// IsExistByField 根据字段判断用户是否存在
func (us *User) IsExistByField(field string, value any) bool {
	err := global.Db.Where(field, value).Find(&us).Error
	if err != nil {
		return false
	}
	if us.ID <= 0 {
		return false
	}
	return true
}

// IfPasswordCorrect 判断密码
func (us *User) IfPasswordCorrect(password string) bool {
	passwordImport := fmt.Sprintf("%s%s%s", us.Salt, password, us.Salt)
	passwordImport = fmt.Sprintf("%x", md5.Sum([]byte(passwordImport)))
	if passwordImport != us.Password {
		return false
	}
	return true
}

// Find 根据id 查询
func (us *User) Find(id uint) {
	_ = global.Db.Where("id", id).Find(&us).Error
}

// FindLiveInfo 查询直播信息
func (us *User) FindLiveInfo(id uint) {
	_ = global.Db.Where("id", id).Preload("LiveInfo").Find(&us).Error
}

func (l *UserList) GetBeLiveList(ids []uint) error {
	return global.Db.Where("id", ids).Preload("LiveInfo").Find(&l).Error
}

func (l *UserList) Search(info common.PageInfo) error {
	return global.Db.Where("`username` LIKE ?", "%"+info.Keyword+"%").Find(&l).Error
}

// GetUsersByQueryTime 查询创建时间比扫表时间晚的用户
func (l *UserList) GetUsersByQueryTime(time string) error {
	return global.Db.Where("created_at > ?", time).Find(&l).Error
}

// GetAllUserIds 获取全部用户
func (l *UserList) GetAllUserIds() error {
	return global.Db.Find(&l).Error
}
