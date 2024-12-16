package cron_events

import (
	"errors"
	"haifengonline/global"
)

type CronEvent struct {
	Id       int64  `json:"id" gorm:"id"`
	LastTime string `json:"last_time" gorm:"last_time"` //上一次扫表的时间
}

func (CronEvent) TableName() string {
	return "lv_cron_events"
}

func (this *CronEvent) GetLastQuery() (*CronEvent, error) {
	err := global.Db.Last(&this).Error
	if err != nil {
		return nil, errors.New("查询lv_cron_events表最后一条记录出错" + err.Error())
	}
	return this, nil
}
