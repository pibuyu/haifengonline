package socket

import (
	"haifengonline/logic/contribution/sokcet"
	liveSocket "haifengonline/logic/live/socket"
	"haifengonline/logic/users/chat"
	"haifengonline/logic/users/chatUser"
	"haifengonline/logic/users/notice"
)

func init() {
	//初始化所有socket
	go liveSocket.Severe.Start()
	go sokcet.Severe.Start()
	go notice.Severe.Start()
	go chat.Severe.Start()
	go chatUser.Severe.Start()
}
