package oss

import (
	"errors"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ice20201109 "github.com/alibabacloud-go/ice-20201109/v2/client"
	sts20150401 "github.com/alibabacloud-go/sts-20150401/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"haifengonline/global"
	"os"
	"strings"
)

var accessKeyId = global.Config.AliyunOss.AccessKeyId

var ossEndpoint = global.Config.AliyunOss.OssEndPoint

var bucketName = global.Config.AliyunOss.Bucket

var accessKeySecret = global.Config.AliyunOss.AccessKeySecret

var roleArn = global.Config.AliyunOss.RoleArn

var roleSessionName = global.Config.AliyunOss.RoleSessionName

var durationSeconds = global.Config.AliyunOss.DurationSeconds

var endpoint = global.Config.AliyunOss.Endpoint

type ConfigStruct struct {
	Expiration string     `json:"expiration"`
	Conditions [][]string `json:"conditions"`
}

type CallbackParam struct {
	CallbackUrl      string `json:"callbackUrl"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
}
type PolicyToken struct {
	AccessKeyId string `json:"access_id"`
	Host        string `json:"host"`
	Expire      int64  `json:"expire"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	Directory   string `json:"dir"`
	Callback    string `json:"callback"`
}

// CreateOSSClient 获取oss文件操作对象
func CreateOSSClient() (*oss.Client, error) {

	os.Setenv("OSS_ACCESS_KEY_ID", accessKeyId)
	os.Setenv("OSS_ACCESS_KEY_SECRET", accessKeySecret)
	//这个方法是从环境变量里读取id和key来生成authorization信息的,所以需要先设置环境变量
	provider, err := oss.NewEnvironmentVariableCredentialsProvider()

	client, err := oss.New(ossEndpoint, accessKeyId, accessKeySecret, oss.SetCredentialsProvider(&provider))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func DeleteOSSFile(filePath []string) error {

	client, err := CreateOSSClient()
	if err != nil {
		return errors.New("创建oss client err : " + err.Error())
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return errors.New("获取bucket err : " + err.Error())
	}

	//删除给定的文件
	for _, path := range filePath {
		//要先预处理一下这个path，因为path除了实际路径还包括src: 这个前缀，以及type:oss这个后缀
		result := strings.Split(path, `"src": "`)
		//split完了之后是一个数组
		result = strings.Split(result[1], `", "type": "oss"`)
		//此时result只剩一个元素，也是我们需要的最终路径
		endPath := result[0]

		global.Logger.Infof("要删除的路径为%s", endPath)
		err := bucket.DeleteObject(endPath)
		if err != nil {
			return errors.New("删除文件 err : " + err.Error())
		}
	}

	return nil
}

func CreateStsClient(accessKeyId *string, accessKeySecret *string) (_result *sts20150401.Client, _err error) {
	config := &openapi.Config{
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
	}
	// 访问的域名
	config.Endpoint = tea.String(endpoint)
	_result = &sts20150401.Client{}
	_result, _err = sts20150401.NewClient(config)
	return _result, _err
}

func GteStsInfo() (*sts20150401.AssumeRoleResponseBodyCredentials, error) {
	client, err := CreateStsClient(tea.String(accessKeyId), tea.String(accessKeySecret))
	if err != nil {
		fmt.Errorf("CreateStsClient方法创建client出错 : " + err.Error())
		return nil, err
	}
	assumeRoleRequest := &sts20150401.AssumeRoleRequest{
		RoleArn:         tea.String(roleArn),
		RoleSessionName: tea.String(roleSessionName),
		//DurationSeconds参数有范围限制的：最小15分钟，最大一个小时，直接设置成一个小时得了
		DurationSeconds: tea.Int64(3600),
	}
	runtime := &util.RuntimeOptions{}
	defer func() {
		if r := tea.Recover(recover()); r != nil {
		}
	}()
	res, err := client.AssumeRoleWithOptions(assumeRoleRequest, runtime)
	if err != nil {
		return nil, err
	}
	if *res.StatusCode != 200 {
		return nil, fmt.Errorf("错误的状态码: %d", res.StatusCode)
	}
	//global.Logger.Infof("打印GteStsInfo方法的返回值:%v", res.Body.Credentials)
	return res.Body.Credentials, nil
}

func CreateIceClient(accessKeyId *string, accessKeySecret *string) (_result *ice20201109.Client, _err error) {
	config := &openapi.Config{
		RegionId:        tea.String("cn-hangzhou"),
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		Endpoint:        tea.String("ice.cn-hangzhou.aliyuncs.com"),
	}
	//result := &ice20201109.Client{}
	result, err := ice20201109.NewClient(config)
	return result, err
}

// RegisterMediaInfo 注册媒体资源
// params:视频的存储路径；媒体资源的类型；和当前时间转成的字符串
func RegisterMediaInfo(inputUrl, mediaType, Title string) (body *ice20201109.RegisterMediaInfoResponseBody, err error) {
	//global.Logger.Infoln("调用了RegisterMediaInfo方法，打印这个信息说明正常调用了，那就可能是后面调用aliyun的client.RegisterMediaInfoWithOptions方法报错了")
	client, err := CreateIceClient(tea.String(accessKeyId), tea.String(accessKeySecret))
	if err != nil {
		global.Logger.Errorf("初始化cilent失败 err : %s", err.Error())
	}
	//global.Logger.Infof("进行媒体资源注册的参数为:URL=%s,mediaType=%s,title=%s", inputUrl, mediaType, Title)

	registerMediaInfoRequest := &ice20201109.RegisterMediaInfoRequest{
		Overwrite: tea.Bool(true),
		InputURL:  tea.String(inputUrl),
		MediaType: tea.String(mediaType),
		Title:     tea.String(Title),
	}

	result, err := client.RegisterMediaInfo(registerMediaInfoRequest)

	if err != nil {
		global.Logger.Errorf("注册媒资失败 err %s ", err.Error())
	}
	if *result.StatusCode != 200 {
		global.Logger.Errorf("注册媒体资源失败，返回的错误信息为:" + err.Error())
		return nil, err
	}
	//global.Logger.Infof("注册媒体资源成功，返回结果为%s，返回的mediaId的长度为:%d", result.Body, len(*result.Body.MediaId))
	return result.Body, nil
}

func GetMediaInfo(mediaID *string) (body *ice20201109.GetMediaInfoResponse, err error) {
	client, err := CreateIceClient(tea.String(accessKeyId), tea.String(accessKeySecret))
	if err != nil {
		global.Logger.Errorf("初始化cilent失败 err : %s", err.Error())
	}

	getMediaInfoRequest := &ice20201109.GetMediaInfoRequest{
		MediaId: mediaID,
	}
	runtime := &util.RuntimeOptions{}
	if r := tea.Recover(recover()); r != nil {
	}
	result, err := client.GetMediaInfoWithOptions(getMediaInfoRequest, runtime)
	global.Logger.Infof("打印iceclient返回的GetMediaInfo的结果：%v", result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SubmitTranscodeJob(taskName, mediaID, outputUrl, template string) (body *ice20201109.SubmitTranscodeJobResponseBody, err error) {
	client, err := CreateIceClient(tea.String(accessKeyId), tea.String(accessKeySecret))
	if err != nil {
		global.Logger.Errorf("初始化cilent失败 err : %s", err.Error())
	}
	inputGroup0 := &ice20201109.SubmitTranscodeJobRequestInputGroup{
		Type:  tea.String("Media"),
		Media: tea.String(mediaID),
	}
	outputGroup0Output := &ice20201109.SubmitTranscodeJobRequestOutputGroupOutput{
		Type:  tea.String("OSS"),
		Media: tea.String(outputUrl),
	}
	outputGroup0ProcessConfigTranscode := &ice20201109.SubmitTranscodeJobRequestOutputGroupProcessConfigTranscode{
		TemplateId: tea.String(template),
	}
	outputGroup0ProcessConfig := &ice20201109.SubmitTranscodeJobRequestOutputGroupProcessConfig{
		Transcode: outputGroup0ProcessConfigTranscode,
	}
	outputGroup0 := &ice20201109.SubmitTranscodeJobRequestOutputGroup{
		ProcessConfig: outputGroup0ProcessConfig,
		Output:        outputGroup0Output,
	}
	submitTranscodeJobRequest := &ice20201109.SubmitTranscodeJobRequest{
		OutputGroup: []*ice20201109.SubmitTranscodeJobRequestOutputGroup{outputGroup0},
		Name:        tea.String(taskName),
		InputGroup:  []*ice20201109.SubmitTranscodeJobRequestInputGroup{inputGroup0},
	}
	runtime := &util.RuntimeOptions{}
	defer func() {
		if r := tea.Recover(recover()); r != nil {
		}
	}()
	result, err := client.SubmitTranscodeJobWithOptions(submitTranscodeJobRequest, runtime)
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func GetTranscodeJob(taskID string) (body *ice20201109.GetTranscodeJobResponseBody, err error) {
	client, err := CreateIceClient(tea.String(accessKeyId), tea.String(accessKeySecret))
	if err != nil {
		global.Logger.Errorf("初始化cilent失败 err : %s", err.Error())
	}
	getTranscodeJobRequest := &ice20201109.GetTranscodeJobRequest{
		ParentJobId: tea.String("9ce776d01f034d23b31bc68ffbb2e276"),
	}
	runtime := &util.RuntimeOptions{}
	result, err := client.GetTranscodeJobWithOptions(getTranscodeJobRequest, runtime)
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}
