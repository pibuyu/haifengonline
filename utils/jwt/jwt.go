package jwt

import (
	"errors"
	"fmt"
	"haifengonline/consts"
	"haifengonline/global"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Hotkey 密钥
var Hotkey = []byte("haifengonline.top")

// SaltStr  密码盐的随机字符串
var SaltStr = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Claims  TOKEN 的结构体
// todo:改造登录模块，用redis+token实现单用户登录
type Claims struct {
	UserID    uint
	LoginTime string
	jwt.StandardClaims
}

// NextToken 就是返回一个token
func NextToken(uid uint) string {
	expireTime := time.Now().Add(7 * 24 * time.Hour)

	claims := &Claims{
		UserID:    uid,
		LoginTime: time.Now().Format(time.DateTime),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(), //过期时间
			IssuedAt:  time.Now().Unix(),
			Issuer:    "root",       // 签名颁发者
			Subject:   "user token", //签名主题
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(Hotkey)
	if err != nil {
		fmt.Println(err)
	}

	//签发token的同时往redis里放一份;然后在中间件查看收到的token和redis中的最新token是否一致
	key := fmt.Sprintf("%d_%s", uid, consts.TokenString)
	global.RedisDb.Set(key, tokenString, 7*24*time.Hour)
	global.Logger.Infof("给用户 %d 签发并向redis投递了token：%s", uid, tokenString)
	return tokenString
}

// ParseToken 解析 Token
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return Hotkey, nil
	})
	if err != nil {
		global.Logger.Errorf("token parse err : " + err.Error())
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

//todo:This is a special comment.获取uid是ctx.getUint("uid")
