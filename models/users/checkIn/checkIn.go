package checkIn

import (
	"haifengonline/global"
	"haifengonline/models/common"
)

type CheckIn struct {
	common.PublicModel
	Uid             uint `json:"uid" gorm:"column:uid"`
	LatestDay       int  `json:"latest_day" gorm:"latest_day"`             //最后一次签到的日期
	ConsecutiveDays int  `json:"consecutive_days" gorm:"consecutive_days"` //连续签到天数
	Integral        int  `json:"integral" gorm:"integral"`                 //签到获得的积分
}

func (CheckIn) TableName() string {
	return "lv_check_in"
}

func GetCheckInRecordByUID(uid uint) *CheckIn {
	c := &CheckIn{}
	global.Db.Where("uid = ?", uid).Find(&c)
	return c

}

func (this *CheckIn) Create() bool {
	if err := global.Db.Create(&this).Error; err != nil {
		return false
	}
	return true
}

func (this *CheckIn) Updates(info map[string]interface{}) error {
	if err := global.Db.Model(&CheckIn{}).Where("uid = ?", this.Uid).Updates(info).Error; err != nil {
		return err
	}
	return nil
}
func (this *CheckIn) Query() bool {
	err := global.Db.Where("uid = ?", this.Uid).Find(this).Error
	if err != nil {
		return false
	}
	return true
}
func (this *CheckIn) Delete() bool {
	err := global.Db.Delete(this).Error
	return err == nil
}

func (this *CheckIn) UpdateColumn(name string, value string, uid uint) error {
	if err := global.Db.Model(&CheckIn{}).Where("uid = ?", uid).UpdateColumn(name, value).Error; err != nil {
		return err
	}
	return nil
}
