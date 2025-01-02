package notice

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"haifengonline/consts"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/socket"
	socketResponse "haifengonline/interaction/response/socket"
	userModel "haifengonline/models/users"
	"haifengonline/models/users/notice"
	noticeModel "haifengonline/models/users/notice"
	"haifengonline/utils/response"
)

type Engine struct {
	//视频房间
	UserMapChannel map[uint]*UserChannel
	Register       chan *UserChannel
	Cancellation   chan *UserChannel
}

type ChanInfo struct {
	Type string
	Data interface{}
}

// UserChannel 用户信息
type UserChannel struct {
	UserInfo *userModel.User
	Socket   *websocket.Conn
	MsgList  chan ChanInfo
}

var Severe = &Engine{
	UserMapChannel: make(map[uint]*UserChannel, 10),
	Register:       make(chan *UserChannel, 10),
	Cancellation:   make(chan *UserChannel, 10),
}

// Start 启动服务
func (e *Engine) Start() {
	for {
		select {
		//注册事件
		case registerMsg := <-e.Register:
			//添加成员和删除成员就是在map里添加和删除用户的id？？
			//添加成员
			e.UserMapChannel[registerMsg.UserInfo.ID] = registerMsg
			registerMsg.NoticeMessage(noticeModel.Online)
		case cancellationMsg := <-e.Cancellation:
			//删除成员
			delete(e.UserMapChannel, cancellationMsg.UserInfo.ID)
		}
	}
}

func CreateNoticeSocket(uid uint, conn *websocket.Conn) (err error) {
	//创建UserChannel
	userChannel := new(UserChannel)
	//绑定ws
	userChannel.Socket = conn
	user := &userModel.User{}
	user.Find(uid)
	userChannel.UserInfo = user
	userChannel.MsgList = make(chan ChanInfo, 20)

	Severe.Register <- userChannel

	go userChannel.Read()
	go userChannel.Writer()
	return nil

}

// Writer 监听写入数据
func (lre *UserChannel) Writer() {
	for {
		select {
		case msg := <-lre.MsgList:
			response.SuccessWs(lre.Socket, msg.Type, msg.Data)
		}
	}
}

// Read 读取数据
func (lre *UserChannel) Read() {
	//链接断开进行离线(用户直接退出了网站，断开ws连接)
	defer func() {
		Severe.Cancellation <- lre
		err := lre.Socket.Close()
		if err != nil {
			return
		}
	}()
	//监听业务通道
	for {
		//检查通达ping通
		lre.Socket.PongHandler()
		_, text, err := lre.Socket.ReadMessage()
		if err != nil {
			return
		}
		info := new(receive.Receive)
		if err = json.Unmarshal(text, info); err != nil {
			response.ErrorWs(lre.Socket, "消息格式错误")
		}
		switch info.Type {

		}
	}
}

func (lre *UserChannel) NoticeMessage(tp string) {
	//获取未读消息
	nl := new(notice.Notice)
	num := nl.GetUnreadNum(lre.UserInfo.ID)
	if num == nil {
		global.Logger.Errorf("通知id为%d用户未读消息失败", lre.UserInfo.ID)
	}
	//global.Logger.Infof("向 %d 用户推送了一条 %s 类型的消息", lre.UserInfo.ID, tp)
	//向这个channel的msgList推送新通知的类型;和未读消息的条数
	lre.MsgList <- ChanInfo{
		Type: consts.NoticeSocketTypeMessage,
		Data: socketResponse.NoticeMessageStruct{
			NoticeType: tp,
			Unread:     num,
		},
	}
}
