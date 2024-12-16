package users

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"haifengonline/logic/users/chat"
	"haifengonline/logic/users/chatUser"
	"haifengonline/logic/users/notice"
	"haifengonline/utils/response"
	"strconv"
)

// NoticeSocket  通知socket
func (us UserControllers) NoticeSocket(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	conn, _ := ctx.Get("conn")
	ws := conn.(*websocket.Conn)
	err := notice.CreateNoticeSocket(uid, ws)
	if err != nil {
		response.ErrorWs(ws, "创建通知socket失败")
	}
}

// ChatSocket  聊天socket
func (us UserControllers) ChatSocket(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	conn, _ := ctx.Get("conn")
	ws := conn.(*websocket.Conn)
	err := chat.CreateChatSocket(uid, ws)
	if err != nil {
		response.ErrorWs(ws, "创建聊天socket失败")
	}
}

func (us UserControllers) ChatByUserSocket(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	conn, _ := ctx.Get("conn")
	//判断是否创建视频socket房间
	tidQuery, _ := strconv.Atoi(ctx.Query("tid"))
	tid := uint(tidQuery)
	ws := conn.(*websocket.Conn)
	err := chatUser.CreateChatByUserSocket(uid, tid, ws)
	if err != nil {
		response.ErrorWs(ws, "创建用户聊天socket失败")
	}
}
