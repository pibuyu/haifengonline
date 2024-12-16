package captcha

import (
	"github.com/dchest/captcha"
	"haifengonline/global"
)

type captchaResponse struct {
	CaptchaId string `json:"captchaId"`
	ImageUrl  string `json:"imageUrl"`
}

func GetCaptcha(id string) (results interface{}, err error) {
	//这里先不用传过来的id试一下
	length := captcha.DefaultLen
	captchaId := captcha.NewLen(length)
	var response = &captchaResponse{
		CaptchaId: captchaId,
		ImageUrl:  "/captcha/" + captchaId + ".png",
	}
	global.Logger.Infof("captcha信息为%s,%s", response.CaptchaId, response.ImageUrl)
	return response, nil
}
