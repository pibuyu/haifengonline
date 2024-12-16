package users

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"haifengonline/consts"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/users"
	response "haifengonline/interaction/response/users"
	"haifengonline/models/common"
	"haifengonline/models/users"
	"haifengonline/models/users/attention"
	"haifengonline/models/users/chat/chatList"
	"haifengonline/models/users/chat/chatMsg"
	"haifengonline/models/users/collect"
	"haifengonline/models/users/favorites"
	"haifengonline/models/users/liveInfo"
	"haifengonline/models/users/notice"
	"haifengonline/models/users/record"
	"haifengonline/utils/conversion"
	"haifengonline/utils/email"
	"haifengonline/utils/jwt"
	"math/rand"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/datatypes"
)

func GetUserInfo(uid uint) (results interface{}, err error) {
	user := new(users.User)
	user.IsExistByField("id", uid)
	//bytes, _ := json.Marshal(user)
	//global.RedisDb.Set(fmt.Sprintf("name_%s", user.Username), string(bytes), 10*time.Minute)
	res := response.UserSetInfoResponse(user)
	return res, nil
}

func SetUserInfo(data *receive.SetUserInfoReceiveStruct, uid uint) (results interface{}, err error) {
	user := &users.User{
		PublicModel: common.PublicModel{ID: uid},
	}
	update := map[string]interface{}{
		"Username":    data.Username,
		"Gender":      data.Gender,
		"BirthDate":   data.BirthDate,
		"IsVisible":   conversion.BoolTurnInt8(*data.IsVisible),
		"Signature":   data.Signature,
		"SocialMedia": data.SocialMedia,
	}

	return user.UpdatePureZero(update), nil
}

func DetermineNameExists(data *receive.DetermineNameExistsStruct, uid uint) (results interface{}, err error) {
	user := new(users.User)
	is := user.IsExistByField("username", data.Username)
	//判断是否未更改
	if user.ID == uid {
		return false, nil
	} else if is {
		return true, nil
	} else {
		return false, nil
	}
}

func UpdateAvatar(data *receive.UpdateAvatarStruct, uid uint) (results interface{}, err error) {
	photo, _ := json.Marshal(common.Img{
		Src: data.ImgUrl,
		Tp:  data.Tp,
	})
	user := &users.User{PublicModel: common.PublicModel{ID: uid}, Photo: photo}
	if user.Update() {
		return conversion.SwitchIngStorageFun(data.Tp, data.ImgUrl)
	} else {
		return nil, fmt.Errorf("更新失败")
	}
}

func GetLiveData(uid uint) (results interface{}, err error) {
	info := new(liveInfo.LiveInfo)
	if info.IsExistByField("uid", uid) {
		results, err = response.GetLiveDataResponse(info)
		if err != nil {
			return nil, fmt.Errorf("获取失败")
		}
		return results, nil
	}
	return common.Img{}, nil
}

func SaveLiveData(data *receive.SaveLiveDataReceiveStruct, uid uint) (results interface{}, err error) {
	img, _ := json.Marshal(common.Img{
		Src: data.ImgUrl,
		Tp:  data.Tp,
	})
	info := &liveInfo.LiveInfo{
		Uid:   uid,
		Title: data.Title,
		Img:   datatypes.JSON(img),
	}
	if info.UpdateInfo() {
		return "修改成功", nil
	} else {
		return nil, fmt.Errorf("修改失败")
	}

}

func SendEmailVerificationCodeByChangePassword(uid uint) (results interface{}, err error) {
	user := new(users.User)
	user.Find(uid)
	//发送方
	mailTo := []string{user.Email}
	// 邮件主题
	code := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(1000000))
	subject := "验证码"
	// 邮件正文
	body := fmt.Sprintf("您正在修改密码,您的验证码为:%s,5分钟有效,请勿转发他人", code)
	err = email.SendMail(mailTo, subject, body)
	if err != nil {
		return nil, err
	}
	err = global.RedisDb.Set(fmt.Sprintf("%s%s", consts.EmailVerificationCodeByChangePassword, user.Email), code, 5*time.Minute).Err()
	if err != nil {
		return nil, err
	}
	return "发送成功", nil

}

func ChangePassword(data *receive.ChangePasswordReceiveStruct, uid uint) (results interface{}, err error) {
	user := new(users.User)
	user.Find(uid)

	if data.Password != data.ConfirmPassword {
		return nil, fmt.Errorf("两次密码不一致！")
	}

	//判断验证码是否正确
	verCode, err := global.RedisDb.Get(fmt.Sprintf("%s%s", consts.EmailVerificationCodeByChangePassword, user.Email)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("验证码过期！")
	}

	if verCode != data.VerificationCode {
		return nil, fmt.Errorf("验证码错误")
	}
	//生成密码盐 8 位 ，salt是随机生成的，每次改密码salt也会改变
	salt := make([]byte, 6)
	for i := range salt {
		salt[i] = jwt.SaltStr[rand.Int63()%int64(len(jwt.SaltStr))]
	}
	password := []byte(fmt.Sprintf("%s%s%s", salt, data.Password, salt))
	passwordMd5 := fmt.Sprintf("%x", md5.Sum(password))

	user.Salt = string(salt)
	user.Password = passwordMd5

	registerRes := user.Update()
	if !registerRes {
		return nil, fmt.Errorf("修改失败")
	}
	return "修改成功", nil
}

func Attention(data *receive.AttentionReceiveStruct, uid uint) (results interface{}, err error) {
	at := new(attention.Attention)
	if at.Attention(uid, data.Uid) {
		if data.Uid == uid {
			return nil, fmt.Errorf("操作失败")
		}
		return "操作成功", nil
	}
	return nil, fmt.Errorf("操作失败")
}

func CreateFavorites(data *receive.CreateFavoritesReceiveStruct, uid uint) (results interface{}, err error) {
	if data.ID == 0 {
		//插入模式
		if len(data.Title) == 0 {
			return nil, fmt.Errorf("标题为空")
		}
		//判断是否只有标题
		if data.ID <= 0 && len(data.Tp) == 0 && len(data.Content) == 0 && len(data.Cover) == 0 {
			//单标题创建
			fs := &favorites.Favorites{Uid: uid, Title: data.Title, Max: 1000}
			if !fs.Create() {
				return nil, fmt.Errorf("创建失败")
			}
			return fmt.Errorf("创建成功"), nil
		} else {
			//资料齐全创建
			cover, _ := json.Marshal(common.Img{
				Src: data.Cover,
				Tp:  data.Tp,
			})
			fs := &favorites.Favorites{
				Uid:     uid,
				Title:   data.Title,
				Content: data.Content,
				Cover:   cover,
				Max:     1000,
			}
			if !fs.Create() {
				return nil, fmt.Errorf("创建失败")
			}
			return fmt.Errorf("创建成功"), nil
		}
	} else {
		//进行更新
		fs := new(favorites.Favorites)
		if !fs.Find(data.ID) {
			return nil, fmt.Errorf("查询失败")
		}
		if fs.Uid != uid {
			return nil, fmt.Errorf("查询非法操作")
		}
		cover, _ := json.Marshal(common.Img{
			Src: data.Cover,
			Tp:  data.Tp,
		})
		fs.Title = data.Title
		fs.Content = data.Content
		fs.Cover = cover
		if !fs.Update() {
			return nil, fmt.Errorf("更新失败")
		}
		return "更新成功", nil
	}
}

func GetFavoritesList(uid uint) (results interface{}, err error) {
	fl := new(favorites.FavoriteList)
	err = fl.GetFavoritesList(uid)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	res, err := response.GetFavoritesListResponse(fl)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DeleteFavorites(data *receive.DeleteFavoritesReceiveStruct, uid uint) (results interface{}, err error) {
	fs := new(favorites.Favorites)
	err = fs.Delete(data.ID, uid)
	if err != nil {
		return nil, err
	}
	return "删除成功", nil
}

func idInIds(id uint, Ids []uint) bool {
	for _, value := range Ids {
		if id == value {
			return true
		}
	}
	return false
}

// FavoriteVideo params：
//
//	data：视频应该位于的收藏夹的ids
func FavoriteVideo(data *receive.FavoriteVideoReceiveStruct, uid uint) (results interface{}, err error) {
	//传递过来的data.IDs可能可能会比原本的ids多，就对应收藏；也可能比之前的少，就对应取消收藏(不会相等，因为前台已经判断过了，ids不变的话不许发请求)
	//step1:取消收藏
	cl := new(collect.CollectsList)
	err = cl.FindVideoExistWhere(data.VideoID) //找一下这个视频原本被哪些收藏夹收藏过
	if err != nil {
		return nil, fmt.Errorf("取消收藏过程,查询所在收藏夹失败")
	}
	ids := make([]uint, 0)  //ids存储的就是视频之前所在的收藏夹了
	for _, v := range *cl { //找出这个视频之前在当前用户的哪些收藏夹里，后续找出差集（对应：取消收藏操作）删除这些差集对应的CollectsList对象
		if v.Uid == uid {
			ids = append(ids, v.FavoritesID)
		}
	}
	//global.Logger.Infof("视频%v之前所在的收藏夹有%v,现在所在的收藏夹有%v", data.VideoID, ids, data.IDs)
	//这个地方的逻辑写反了，应该是遍历之前的ids，如果不在现在的data.IDs里，说明应该取消收藏
	for _, id := range ids {
		if !idInIds(id, data.IDs) { //现在所在的收藏夹id如果不在之前所在的收藏夹的ids中，说明这个收藏被取消了
			//global.Logger.Infof("视频%v不在id为%v收藏夹里了", data.VideoID, id)
			oneCollect := new(collect.Collect)
			err := global.Db.Debug().Where("uid", uid).Where("favorites_id", id).Where("video_id", data.VideoID).Delete(oneCollect).Error
			if err != nil {
				return nil, fmt.Errorf("取消收藏过程,取消收藏失败")
			}
		}
	}

	//step2:新增收藏项
	for _, id := range data.IDs {
		fs := new(favorites.Favorites)
		fs.Find(id)

		if fs.Uid != uid {
			return nil, fmt.Errorf("非法操作")
		}
		if len(fs.CollectList)+1 > fs.Max {
			return nil, fmt.Errorf("收藏夹已满")
		}

		//判断是否重复收藏
		oneCollect := new(collect.Collect)
		err := global.Db.Where("uid", uid).Where("video_id", data.VideoID).Where("favorites_id", id).Find(oneCollect).Error
		if err != nil {
			return nil, fmt.Errorf("查重失败")
		}
		if oneCollect.ID > 0 { //说明已经被当前用户的当前收藏夹收藏过了
			continue
		}

		cl := &collect.Collect{
			Uid:         uid,
			FavoritesID: id,
			VideoID:     data.VideoID,
		}
		if !cl.Create() {
			return nil, fmt.Errorf("收藏失败")
		}
	}

	return "操作成功", nil
}

func GetFavoritesListByFavoriteVideo(data *receive.GetFavoritesListByFavoriteVideoReceiveStruct, uid uint) (results interface{}, err error) {
	//获取收藏夹列表
	fl := new(favorites.FavoriteList)
	err = fl.GetFavoritesList(uid) //获取指定用户创建的所有收藏夹，关联的有：收藏夹与被收藏视频的信息、用户信息
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	//查询该视频在哪些收藏夹内已收藏，即存在于哪个收藏夹内，返回的是这个收藏夹的uid、FavoritesID和videoID及其关联表
	cl := new(collect.CollectsList)
	err = cl.FindVideoExistWhere(data.VideoID)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	ids := make([]uint, 0)
	for _, v := range *cl { //cl是Collect的列表；可能存在于多个收藏夹内，返回这些收藏夹的ids
		ids = append(ids, v.FavoritesID)
	}

	res, err := response.GetFavoritesListByFavoriteVideoResponse(fl, ids)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetFavoriteVideoList(data *receive.GetFavoriteVideoListReceiveStruct) (results interface{}, err error) {
	cl := new(collect.CollectsList)
	err = cl.GetVideoInfo(data.FavoriteID)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	res, err := response.GetFavoriteVideoListResponse(cl)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetCollectListName 根据收藏夹id获取收藏夹的title
func GetCollectListName(data *receive.GetCollectListNameReceiveStruct) (results interface{}, err error) {
	favorite := new(favorites.Favorites)
	global.Db.Model(&favorite).Where("id = ?", data.FavoriteID).Find(results)
	return results, nil
}

func GetRecordList(data *receive.GetRecordListReceiveStruct, uid uint) (results interface{}, err error) {
	rl := new(record.RecordsList)
	err = rl.GetRecordListByUid(uid, data.PageInfo)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	res, err := response.GetRecordListResponse(rl)
	if err != nil {
		return nil, fmt.Errorf("响应失败")
	}
	return res, nil
}

func ClearRecord(uid uint) (results interface{}, err error) {
	rl := new(record.Record)
	err = rl.ClearRecord(uid)
	if err != nil {
		return nil, fmt.Errorf("清空失败")
	}
	return "清空完成", nil
}

func DeleteRecordByID(data *receive.DeleteRecordByIDReceiveStruct, uid uint) (results interface{}, err error) {
	rl := new(record.Record)
	err = rl.DeleteRecordByID(data.ID, uid)
	if err != nil {
		return nil, fmt.Errorf("删除失败")
	}
	return "删除成功", nil
}

func GetNoticeList(data *receive.GetNoticeListReceiveStruct, uid uint) (results interface{}, err error) {
	//获取用户通知
	messageType := make([]string, 0)
	nl := new(notice.NoticesList)
	switch data.Type {
	case "comment":
		messageType = append(messageType, notice.VideoComment, notice.ArticleComment)
		break
	case "like":
		messageType = append(messageType, notice.VideoLike, notice.ArticleLike)
	//todo:系统通知的类型添加在这里
	case "system":
		messageType = append(messageType, notice.UserLogin, notice.DailyReport, notice.DailyReport)
	}

	err = nl.GetNoticeList(data.PageInfo, messageType, uid)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	//记录全部已读
	n := new(notice.Notice)
	err = n.ReadAll(uid)
	if err != nil {
		return nil, fmt.Errorf("设置通知消息为已读失败")
	}
	res, err := response.GetNoticeListResponse(nl)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetChatList(uid uint) (results interface{}, err error) {
	//获取消息列表
	cList := new(chatList.ChatList)
	err = cList.GetListByIO(uid)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	ids := make([]uint, 0)
	for _, v := range *cList {
		ids = append(ids, v.Tid)
	}
	msgList := make(map[uint]*chatMsg.MsgList, 0)
	for _, v := range ids {
		ml := new(chatMsg.MsgList)
		err = ml.FindList(uid, v)
		if err != nil {
			break
		}
		msgList[v] = ml
	}
	res, err := response.GetChatListResponse(cList, msgList)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetChatHistoryMsg(data *receive.GetChatHistoryMsgStruct, uid uint) (results interface{}, err error) {
	//查询历史消息
	cm := new(chatMsg.MsgList)
	err = cm.FindHistoryMsg(uid, data.Tid, data.LastTime)
	if err != nil {
		return nil, fmt.Errorf("查询失败")
	}
	res, err := response.GetChatHistoryMsgResponse(cm)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func PersonalLetter(data *receive.PersonalLetterReceiveStruct, uid uint) (results interface{}, err error) {
	cm := new(chatMsg.Msg)
	err = cm.GetLastMessage(uid, data.ID)
	if err != nil {
		return nil, fmt.Errorf("操作失败")
	}
	var lastTime time.Time
	if cm.ID > 0 {
		lastTime = cm.CreatedAt
	} else {
		lastTime = time.Now()
	}
	ci := &chatList.ChatsListInfo{
		Uid:         uid,
		Tid:         data.ID,
		LastMessage: cm.Message,
		LastAt:      lastTime,
	}
	err = ci.AddChat()
	if err != nil {
		return nil, fmt.Errorf("操作失败")
	}
	return "操作成功", nil
}

func DeleteChatItem(data *receive.DeleteChatItemReceiveStruct, uid uint) (results interface{}, err error) {
	ci := new(chatList.ChatsListInfo)
	err = ci.DeleteChat(data.ID, uid)
	if err != nil {
		return nil, fmt.Errorf("删除失败")
	}
	return "操作成功", nil
}

//func CheckIn(uid uint) (results interface{}, err error) {
//	check := new(checkIn.CheckIn)
//	check.Uid = uid
//
//	return "签到成功", nil
//}
