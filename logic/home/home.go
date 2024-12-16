package home

import (
	"fmt"
	"haifengonline/consts"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/home"
	response "haifengonline/interaction/response/home"
	"haifengonline/models/contribution/video"
	"haifengonline/models/home/rotograph"
	"haifengonline/utils/email"
	"haifengonline/utils/validator"
)

func GetHomeInfo(data *receive.GetHomeInfoReceiveStruct) (results interface{}, err error) {
	//获取主页轮播图
	rotographList := new(rotograph.List)
	err = rotographList.GetAll()
	if err != nil {
		return nil, err
	}

	//获取主页推荐视频
	videoList := new(video.VideosContributionList)
	//todo：这里明显出现了缓存和数据库不一致的情况，因为我把数据库的视频删除之后，缓存中仍然保存了这些数据，返回给前台，导致读取不到视频
	// todo:这里应该先去redis的zset里找一下有没有符合要求的，比方说我们按照heat作为score，返回热度前10的视频
	//要怎么组织videoList的信息，放在zset里？考虑向zset里放热门视频的全部信息，直接返回
	// todo:前端每次刷新主页都会将旧数据和新数据拼接在一起，因此只有初始化主页，即pageInfo信息中page=1时才走缓存，不然会重复请求缓存中的数据
	//if data.PageInfo.Page == 1 {
	//	global.Logger.Infoln("首页初始化走的是缓存")
	//	result, err := global.RedisDb.ZRevRange(consts.HeatestVideo, 0, 14).Result()
	//	if err != nil {
	//		//redis出错就去查数据库，不应该返回err
	//		global.Logger.Errorln("zset查询热门视频出错")
	//	}
	//	//走缓存
	//	if len(result) != 0 {
	//		var err2 error
	//		global.Logger.Infof("zset里取出的数据为：%v", result)
	//		for _, value := range result {
	//			//将取出来的内容转化为VideosContribution类型，放进videoList
	//			var videoContro video.VideosContribution
	//			err2 = json.Unmarshal([]byte(value), &videoContro) //这里传进去的第二个参数必须是指针，也就是&videoContro，否则会报non-pointer错误
	//			if err2 != nil {
	//				global.Logger.Errorln("json反序列化VideosContribution出错：" + err.Error())
	//				break
	//			}
	//			*videoList = append(*videoList, videoContro)
	//		}
	//		//反序列化不出错才能在这里返回，否则还是走数据库
	//		if err2 == nil {
	//			res := &response.GetHomeInfoResponse{}
	//			res.Response(rotographList, videoList)
	//			return res, nil
	//		}
	//	}
	//}
	//走数据库，同时更新缓存
	err = videoList.GetHoneVideoList(data.PageInfo)
	if err != nil {
		return nil, err
	}
	res := &response.GetHomeInfoResponse{}
	res.Response(rotographList, videoList)

	//转化为redis.Z类型存储到zset
	//for _, video := range *videoList {
	//	videoJson, _ := json.Marshal(video)
	//	Z := redis.Z{
	//		Score:  float64(video.Heat),
	//		Member: videoJson,
	//	}
	//	global.RedisDb.ZAdd(consts.HeatestVideo, Z)
	//}

	return res, nil
}

func SubmitBug(data *receive.SubmitBugReceiveStruct) (results interface{}, err error) {
	//_, err = global.Producer.WriteMessages(kafka.Message{Value: []byte(data.Content)})
	//if err != nil {
	//	global.Logger.Errorf("生产者写入消息失败：%v", err)
	//}

	//其实就是给小号发个邮件
	emailTo := []string{consts.SystemEmail}
	err = email.SendMail(emailTo, "用户反馈的bug信息", fmt.Sprintf("用户反馈的bug信息为：%s\n用户留下的联系方式为:%s", data.Content, data.Phone))
	if err != nil {
		global.Logger.Errorf("向系统邮箱发送bug反馈信息失败：%v", err)
		return "保存bug信息失败", err
	}
	//然后看有没有留下邮箱，给对方发送一个反馈邮件
	phone := data.Phone
	if validator.VerifyMobileFormat(phone) {
		//现在不支持给人回短信
	}
	if validator.VerifyEmailFormat(phone) {
		//可以给人回个邮件答复一下
		if err = email.SendMail([]string{data.Phone}, "感谢您的反馈", "已收到您反馈的问题，处理完成后会给您答复"); err != nil {
			global.Logger.Errorf("给用户反馈邮件失败:%v", err)
		}
		global.Logger.Infoln("成功给用户推送确认信息邮件")
	}
	return "success", nil
}
