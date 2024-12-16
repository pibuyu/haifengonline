package checkin

import (
	"errors"
	"fmt"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/users"
	"haifengonline/models/users/checkIn"
	"haifengonline/utils/date"
	"time"
)

func CheckIn(data *receive.CheckInRequestStruct) (results interface{}, err error) {

	//先查询是否有过历史记录
	//global.Logger.Infoln("checkin方法调用")
	check := &checkIn.CheckIn{
		Uid: data.UID,
	}
	if !check.Query() {
		return nil, errors.New("查询签到历史记录出错")
	}

	if check.LatestDay == 0 {
		//那就创建一个记录
		check.Uid = data.UID
		check.LatestDay = date.GetDay(time.Now())
		check.ConsecutiveDays = 1
		if !check.Create() {
			return nil, errors.New("创建签到记录出错")
		}
	} else {
		//判断是否为连续签到：latestDay是否为昨天
		if check.LatestDay == date.GetDay(time.Now()) {
			//这个地方不能给前端返回error，不算错误，只是一种状态，应该返回code：200和提示信息
			return "今天已经签到过了，请勿重复签到", nil
		}
		//是连续签到
		if check.LatestDay == date.GetYesterday() {
			check.ConsecutiveDays += 1
		} else {
			//不是连续签到
			check.ConsecutiveDays = 1
		}
		check.Integral += 1
		if err := check.Updates(map[string]interface{}{"consecutive_days": check.ConsecutiveDays, "integral": check.Integral, "latest_day": date.GetDay(time.Now())}); err != nil {
			return nil, err
		}
	}
	global.Logger.Infof("用户%d成功签到，已经连续签到%d天", check.Uid, check.ConsecutiveDays)
	return fmt.Sprintf("签到成功，您已连续签到%d天", check.ConsecutiveDays), nil
}
