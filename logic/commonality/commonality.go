package commonality

import (
	"fmt"
	"haifengonline/global"
	receive "haifengonline/interaction/receive/commonality"
	response "haifengonline/interaction/response/commonality"
	"haifengonline/models/contribution/video"
	"haifengonline/models/sundry/upload"
	"haifengonline/models/users"
	"haifengonline/models/users/attention"
	"haifengonline/utils/conversion"
	"haifengonline/utils/location"
	"haifengonline/utils/oss"
	"haifengonline/utils/validator"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	//Temporary 文件文件存续位置
	Temporary = filepath.ToSlash("assets/tmp")
)

// OssSTS 这个方法返回500的原因是config.ini配置有问题
func OssSTS() (results interface{}, err error) {
	info, err := oss.GteStsInfo()
	if err != nil {
		global.Logger.Errorf("获取OssSts密钥失败 错误原因 :%s", err.Error())
		return nil, fmt.Errorf("获取失败")
	}
	res, err := response.GteStsInfo(info)
	if err != nil {
		return nil, fmt.Errorf("响应失败")
	}
	//global.Logger.Infof("返回给前端的ossSts信息为%v", res)
	return res, nil
}

func Upload(file *multipart.FileHeader, ctx *gin.Context) (results interface{}, err error) {
	//如果文件大小超过maxMemory,则使用临时文件来存储multipart/form中文件数据
	err = ctx.Request.ParseMultipartForm(128)
	if err != nil {
		return
	}
	mForm := ctx.Request.MultipartForm
	//上传文件名
	var fileName string
	fileName = strings.Join(mForm.Value["name"], fileName)
	var fileInterface string
	fileInterface = strings.Join(mForm.Value["interface"], fileInterface)

	method := new(upload.Upload)
	if !method.IsExistByField("interface", fileInterface) {
		return nil, fmt.Errorf("上传接口不存在")
	}
	if len(method.Path) == 0 {
		return nil, fmt.Errorf("请联系管理员设置接口保存路径")
	}
	index := strings.LastIndex(fileName, ".")
	suffix := fileName[index:]
	err = validator.CheckVideoSuffix(suffix)
	if err != nil {
		return nil, fmt.Errorf("非法后缀！")
	}
	if !location.IsDir(method.Path) {
		// 创建多级目录
		if err = os.MkdirAll(method.Path, 0775); err != nil {
			global.Logger.Errorf("创建文件报错路径失败 创建路径为：%s 错误原因 : %s", method.Path, err.Error())
			return nil, fmt.Errorf("创建保存路径失败")
		}
	}
	dst := filepath.ToSlash(method.Path + "/" + fileName)
	err = ctx.SaveUploadedFile(file, dst)
	if err != nil {
		global.Logger.Errorf("保存文件失败-保存路径为：%s ,错误原因 : %s", dst, err.Error())
		return nil, fmt.Errorf("上传失败")
	} else {
		return dst, nil
	}
}

func UploadSlice(file *multipart.FileHeader, ctx *gin.Context) (results interface{}, err error) {
	//如果文件大小超过maxMemory,则使用临时文件来存储multipart/form中文件数据
	err = ctx.Request.ParseMultipartForm(128)
	if err != nil {
		return
	}
	mForm := ctx.Request.MultipartForm
	//上传文件名
	var fileName string
	fileName = strings.Join(mForm.Value["name"], fileName)
	var fileInterface string
	fileInterface = strings.Join(mForm.Value["interface"], fileInterface)

	method := new(upload.Upload)
	if !method.IsExistByField("interface", fileInterface) {
		return nil, fmt.Errorf("上传接口不存在")
	}
	if len(method.Path) == 0 {
		return nil, fmt.Errorf("请联系管理员设置接口保存路径")
	}
	if !location.IsDir(Temporary) {
		if err = os.MkdirAll(Temporary, 0775); err != nil {
			global.Logger.Errorf("创建临时文件报错路径失败 创建路径为：%s", method.Path)
			return nil, fmt.Errorf("创建保存路径失败")
		}
	}
	dst := filepath.ToSlash(Temporary + "/" + fileName)
	err = ctx.SaveUploadedFile(file, dst)
	if err != nil {
		global.Logger.Errorf("分片上传保存失败-保存路径为：%s ,错误原因 : %s ", dst, err.Error())
		return nil, fmt.Errorf("上传失败")
	} else {
		_ = os.Chmod(dst, 0775)
		return dst, nil
	}
}

func UploadCheck(data *receive.UploadCheckStruct) (results interface{}, err error) {
	method := new(upload.Upload)
	if !method.IsExistByField("interface", data.Interface) {
		return nil, fmt.Errorf("未配置上传方法")
	}
	list := make(receive.UploadSliceList, 0)
	path := filepath.ToSlash(method.Path + "/" + data.FileMd5)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		//文件已存在
		global.Logger.Infof("上传文件 %s 已存在", data.FileMd5)
		return response.UploadCheckResponse(true, list, path)
	}
	//取出未上传的分片
	for _, v := range data.SliceList {
		if _, err := os.Stat(filepath.ToSlash(Temporary + "/" + v.Hash)); os.IsNotExist(err) {
			list = append(list, receive.UploadSliceInfo{
				Index: v.Index,
				Hash:  v.Hash,
			})
		}
	}
	return response.UploadCheckResponse(false, list, "")
}

func UploadMerge(data *receive.UploadMergeStruct) (results interface{}, err error) {
	method := new(upload.Upload)
	if !method.IsExistByField("interface", data.Interface) {
		return nil, fmt.Errorf("未配置上传方法")
	}
	if !location.IsDir(filepath.ToSlash(method.Path)) {
		if err = os.MkdirAll(filepath.ToSlash(method.Path), 0775); err != nil {
			global.Logger.Errorf("创建文件报错路径失败 创建路径为：%s", method.Path)
			return nil, fmt.Errorf("创建保存路径失败")
		}
	}
	dst := filepath.ToSlash(method.Path + "/" + data.FileName)
	list := make(receive.UploadSliceList, 0)
	path := filepath.ToSlash(method.Path + "/" + data.FileName)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		//文件已存在直接返回
		return dst, nil
	}
	//取出未上传的分片
	for _, v := range data.SliceList {
		if _, err := os.Stat(filepath.ToSlash(Temporary + "/" + v.Hash)); os.IsNotExist(err) {
			list = append(list, receive.UploadSliceInfo{
				Index: v.Index,
				Hash:  v.Hash,
			})
		}
	}
	if len(list) > 0 {
		global.Logger.Warnf("上传文件 %s 分片未全部上传", data.FileName)
		return nil, fmt.Errorf("分片未全部上传")
	}
	cf, err := os.Create(dst)
	if err != nil {
		global.Logger.Errorf("创建的合并后文件失败 err : %s", err)
	}
	if err := cf.Close(); err != nil {
		global.Logger.Errorf("创建的合并后文件释放内存失败 %d", err)
	}
	fileInfo, err := os.OpenFile(dst, os.O_APPEND|os.O_WRONLY, os.ModeSetuid)
	if err != nil {
		global.Logger.Errorf("打开创建的合并后文件失败  path %s err : %s", dst, err)
	}
	defer func(fileInfo *os.File) {
		if err := fileInfo.Close(); err != nil {
			global.Logger.Errorf("关闭资源 err : %s", err)
		}
	}(fileInfo)
	//合并操作
	for _, v := range data.SliceList {
		tmpFile, err := os.OpenFile(filepath.ToSlash(Temporary+"/"+v.Hash), os.O_RDONLY, os.ModePerm)
		if err != nil {
			global.Logger.Errorf("合并操作打开临时分片失败 错误原因 : %s", err)
			break
		}
		b, err := ioutil.ReadAll(tmpFile)
		if err != nil {
			global.Logger.Errorf("合并操作读取分片失败 错误原因 : %s", err)
			break
		}
		if _, err := fileInfo.Write(b); err != nil {
			global.Logger.Errorf("合并分片追加错误 错误原因 : %s", err)
			return nil, fmt.Errorf("合并分片追加错误")
			break
		}
		// 关闭分片
		if err := tmpFile.Close(); err != nil {
			global.Logger.Errorf("关闭分片错误 错误原因 : %s", err)
		}
		if err := os.Remove(tmpFile.Name()); err != nil {
			global.Logger.Errorf("合并操作删除临时分片失败 错误原因 : %s", err)
		}
	}
	return dst, nil
}

func UploadingMethod(data *receive.UploadingMethodStruct) (results interface{}, err error) {
	method := new(upload.Upload)
	if method.IsExistByField("interface", data.Method) {
		return response.UploadingMethodResponse(method.Method), nil
	} else {
		return nil, fmt.Errorf("未配置上传方法")
	}
}

func UploadingDir(data *receive.UploadingDirStruct) (results interface{}, err error) {
	method := new(upload.Upload)
	if method.IsExistByField("interface", data.Interface) {
		return response.UploadingDirResponse(method.Path, method.Quality), nil
	} else {
		return nil, fmt.Errorf("未配置上传方法")
	}
}

func GetFullPathOfImage(data *receive.GetFullPathOfImageMethodStruct) (results interface{}, err error) {
	path, err := conversion.SwitchIngStorageFun(data.Type, data.Path)
	if err != nil {
		return nil, err
	}
	return path, nil
}

func Search(data *receive.SearchStruct, uid uint) (results interface{}, err error) {
	switch data.Type {
	case "video":
		//视频搜索
		list := new(video.VideosContributionList)
		err = list.Search(data.PageInfo)
		if err != nil {
			return nil, fmt.Errorf("查询失败")
		}
		res, err := response.SearchVideoResponse(list)
		if err != nil {
			return nil, fmt.Errorf("响应失败")
		}
		return res, nil
		break
	case "user":
		list := new(users.UserList)
		err := list.Search(data.PageInfo)
		if err != nil {
			return nil, fmt.Errorf("查询失败")
		}
		aids := make([]uint, 0)
		if uid != 0 {
			//用户登入情况下
			al := new(attention.AttentionsList)
			err = al.GetAttentionList(uid)
			if err != nil {
				global.Logger.Errorf("用户id %d 获取取关注列表失败,错误原因 : %s ", uid, err.Error())
				return nil, fmt.Errorf("获取关注列表失败")
			}
			for _, v := range *al {
				aids = append(aids, v.AttentionID)
			}
		}
		res, err := response.SearchUserResponse(list, aids)
		return res, nil
		break
	default:
		return nil, fmt.Errorf("未匹配的类型")
	}
	return
}

func RegisterMedia(data *receive.RegisterMediaStruct) (results interface{}, err error) {
	path, _ := conversion.SwitchIngStorageFun(data.Type, data.Path)
	//注册媒资

	registerMediaBody, err := oss.RegisterMediaInfo(path, "video", time.Now().String())
	if err != nil {
		return nil, fmt.Errorf("注册媒资失败")
	}
	//global.Logger.Infof("注册媒体资源的返回结果：%v", registerMediaBody)
	if registerMediaBody == nil {
		global.Logger.Infoln("注册媒体资源的返回结果为空")
	}
	return registerMediaBody.MediaId, nil
}
