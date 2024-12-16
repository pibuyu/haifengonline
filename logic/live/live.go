package live

import (
	"encoding/json"
	"fmt"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/live"
	response "haifengonline/interaction/response/live"
	"haifengonline/models/users"
	"haifengonline/models/users/record"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func GetLiveRoom(uid uint) (results interface{}, err error) {
	//请求直播服务器
	//下面这个url实际上就是http://127.0.0.1:8090/control/get?room=${uid}
	url := global.Config.LiveConfig.Agreement + "://" + global.Config.LiveConfig.IP + ":" + global.Config.LiveConfig.Api + "/control/get?room="
	url = url + strconv.Itoa(int(uid))
	// 创建http get请求
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	// 解析http请求中body 数据到我们定义的结构体中
	ReqGetRoom := new(receive.ReqGetRoom)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(ReqGetRoom); err != nil {
		return nil, err
	}
	if ReqGetRoom.Status != 200 {
		return nil, fmt.Errorf("获取直播地址失败")
	}
	return response.GetLiveRoomResponse("rtmp://"+global.Config.LiveConfig.IP+":"+global.Config.LiveConfig.RTMP+"/live", ReqGetRoom.Data), nil
}

func GetLiveRoomInfo(data *receive.GetLiveRoomInfoReceiveStruct, uid uint) (results interface{}, err error) {
	userInfo := new(users.User)
	userInfo.FindLiveInfo(data.RoomID)
	//http://8.138.149.242:7001/live/37.flv
	flv := global.Config.LiveConfig.Agreement + "://" + global.Config.LiveConfig.IP + ":" + global.Config.LiveConfig.FLV + "/live/" + strconv.Itoa(int(data.RoomID)) + ".flv"

	if uid > 0 {
		//添加历史记录
		rd := new(record.Record)
		err = rd.AddLiveRecord(uid, data.RoomID)
		if err != nil {
			return nil, fmt.Errorf("添加历史记录失败")
		}
	}
	return response.GetLiveRoomInfoResponse(userInfo, flv), nil
}

func GetBeLiveList() (results interface{}, err error) {
	//取开通播放用户id
	//http://8.138.149.242:8090/stat/livestat
	url := global.Config.LiveConfig.Agreement + "://" + global.Config.LiveConfig.IP + ":" + global.Config.LiveConfig.Api + "/stat/livestat"

	resp, err := http.Get(url)
	global.Logger.Infof("获取直播在线list的请求url为%s,请求返回结果为%v", url, resp)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	// 解析http请求中body 数据到我们定义的结构体中
	livestat := new(receive.LivestatRes)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(livestat); err != nil {
		return nil, fmt.Errorf("解析失败")
	}
	if livestat.Status != 200 {
		return nil, fmt.Errorf("获取直播列表失败")
	}
	//获取live中正在直播的列表
	keys := make([]uint, 0)
	for _, kv := range livestat.Data.Publishers {
		ka := strings.Split(kv.Key, "live/")
		uintKey, _ := strconv.ParseUint(ka[1], 10, 19)
		keys = append(keys, uint(uintKey))
	}
	global.Logger.Infof("查询userList的keys为%v", keys)
	userList := new(users.UserList)
	if len(keys) > 0 {
		err = userList.GetBeLiveList(keys)
		if err != nil {
			return nil, fmt.Errorf("查询失败")
		}
	}
	return response.GetBeLiveListResponse(userList), nil
}
