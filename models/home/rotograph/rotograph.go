package rotograph

import (
	"gorm.io/datatypes"
	"haifengonline/global"
	"haifengonline/models/common"
)

type Rotograph struct {
	common.PublicModel
	Title string         `json:"title" gorm:"column:title"`
	Cover datatypes.JSON `json:"cover" gorm:"column:cover"`
	Color string         `json:"color" gorm:"column:color" `
	Type  string         `json:"type" gorm:"column:type"`
	ToId  uint           `json:"to_id" gorm:"column:to_id"`
}

type List []Rotograph

func (Rotograph) TableName() string {
	return "lv_home_rotograph"
}

func (l *List) GetAll() error {
	err := global.Db.Find(&l).Error
	return err
}

func (r *Rotograph) Create() error {
	return global.Db.Create(r).Error
}
