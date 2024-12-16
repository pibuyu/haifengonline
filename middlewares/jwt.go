package middlewares

import (
	"fmt"
	"github.com/gorilla/websocket"
	"haifengonline/consts"
	"haifengonline/global"
	"haifengonline/models/users"
	"haifengonline/utils/jwt"
	ControllersCommon "haifengonline/utils/response"
	Response "haifengonline/utils/response"
	"haifengonline/utils/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// VerificationToken 请求头中携带token
func VerificationToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")

		//验证是否为redis中的最新token；如果不是，就踹下线
		//todo：现在逻辑大体是对的，但是只有在访问需要登陆的接口时，才会把旧的连接踹掉，需要考虑怎么立刻把旧的连接踹掉

		//token为空直接重定向，就不会报：“请求错误”的提示信息了
		if len(token) == 0 {
			ControllersCommon.NotLogin(c, "未登录,token为空！")
			c.Abort()
			return
		}
		claim, err := jwt.ParseToken(token)

		key := fmt.Sprintf("%d_%s", claim.UserID, consts.TokenString)
		redisToken, _ := global.RedisDb.Get(key).Result()
		if redisToken != token {
			ControllersCommon.NotLogin(c, "您已在别处登录！")
			c.Abort()
			return
		}
		if err != nil {
			ControllersCommon.NotLogin(c, "Token过期")
			c.Abort()
			return
		}
		u := new(users.User)
		if !u.IsExistByField("id", claim.UserID) {
			//没有改用户的情况下
			ControllersCommon.NotLogin(c, "用户异常")
			c.Abort()
			return
		}
		//在这里用jwt将uid放进了ctx里面
		//todo:需要在这里进行权限验证，避免水平越权和垂直越权问题
		c.Set("uid", u.ID)
		c.Set("currentUserName", u.Username)
		c.Next()
	}
}

// VerificationTokenAsParameter body参数中携带token
func VerificationTokenAsParameter() gin.HandlerFunc {
	type qu struct {
		Token string `json:"token"`
	}
	return func(c *gin.Context) {
		req := new(qu)
		if err := c.ShouldBindBodyWith(req, binding.JSON); err != nil {
			validator.CheckParams(c, err)
			return
		}
		token := req.Token
		claim, err := jwt.ParseToken(token)
		if err != nil {
			ControllersCommon.NotLogin(c, "Token过期")
			c.Abort()
			return
		}
		u := new(users.User)
		if !u.IsExistByField("id", claim.UserID) {
			//没有改用户的情况下
			ControllersCommon.NotLogin(c, "用户异常")
			c.Abort()
			return
		}
		c.Set("uid", u.ID)
		c.Set("currentUserName", u.Username)
		c.Next()
	}
}

// VerificationTokenNotNecessary 请求头中携带token (非必须)
func VerificationTokenNotNecessary() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if len(token) == 0 {
			//用户未登入时不验证
			c.Next()
		} else {
			//用户登入情况
			claim, err := jwt.ParseToken(token)
			if err != nil {
				c.Next()
			}
			u := new(users.User)
			if !u.IsExistByField("id", claim.UserID) {
				//没有改用户的情况下
				ControllersCommon.NotLogin(c, "用户异常")
				c.Abort()
				return
			}
			c.Set("uid", u.ID)
			c.Set("currentUserName", u.Username)
			c.Next()
		}
	}
}

func VerificationTokenAsSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		//升级ws 以便返回消息
		conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			http.NotFound(c.Writer, c.Request)
			c.Abort()
			return
		}
		token := c.Query("token")
		claim, err := jwt.ParseToken(token)
		if err != nil {
			Response.NotLoginWs(conn, "Token 验证失败")
			_ = conn.Close()
			c.Abort()
			return
		}
		u := new(users.User)
		if !u.IsExistByField("id", claim.UserID) {
			//没有改用户的情况下
			Response.NotLoginWs(conn, "用户异常")
			_ = conn.Close()
			c.Abort()
			return
		}
		c.Set("uid", u.ID)
		c.Set("conn", conn)
		c.Set("currentUserName", u.Username)
		c.Next()
	}
}
