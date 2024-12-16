package checkin

import (
	"errors"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/users"
	"haifengonline/models/users/checkIn"
)

func GetUserIntegral(data *receive.GetUserIntegralRequest) (results interface{}, err error) {
	//global.Logger.Infoln("进入到GetUserIntegral函数")
	c := &checkIn.CheckIn{}
	if err := global.Db.Model(&checkIn.CheckIn{}).Where("uid = ?", data.UID).Find(&c).Error; err != nil {
		return nil, errors.New("查询积分出错")
	}

	return c.Integral, nil
}
