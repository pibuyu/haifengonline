package cron_events

import (
	"haifengonline/global"
	"haifengonline/models/common"
)

// RuntimeLogEntry 日志的结构体
type RuntimeLogEntry struct {
	common.PublicModel
	Time     string `json:"time"`
	Level    string `json:"level"`
	Msg      string `json:"msg"`
	File     string `json:"file"` //避免解析info日志出现null值
	Function string `json:"function"`
}

func (RuntimeLogEntry) TableName() string {
	return "lv_runtime_log"

}

func (this *RuntimeLogEntry) Create() error {
	return global.Db.Create(&this).Error
}
