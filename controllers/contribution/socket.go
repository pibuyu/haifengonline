package contribution

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"haifengonline/logic/contribution/sokcet"
	"haifengonline/utils/response"
	"strconv"
)

// VideoSocket  观看视频建立的socket
func (c Controllers) VideoSocket(ctx *gin.Context) {
	uid := ctx.GetUint("uid")
	conn, _ := ctx.Get("conn")
	ws := conn.(*websocket.Conn)
	//判断是否创建视频socket房间
	id, _ := strconv.Atoi(ctx.Query("videoID"))
	videoID := uint(id)
	//无人观看主动创建
	if sokcet.Severe.VideoRoom[videoID] == nil {
		sokcet.Severe.VideoRoom[videoID] = make(sokcet.UserMapChannel, 10)
	}
	err := sokcet.CreateVideoSocket(uid, videoID, ws)
	if err != nil {
		response.ErrorWs(ws, "创建socket失败")
	}
}
