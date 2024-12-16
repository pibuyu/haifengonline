package crons

import (
	"bufio"
	"haifengonline/Init/kafkaConsumer"
	"haifengonline/consts"
	"haifengonline/global"
	"haifengonline/models/cron_events"
	"haifengonline/models/users/attention"

	"encoding/json"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	userModel "haifengonline/models/users"
	noticeModel "haifengonline/models/users/notice"
	"log"
	"os"
	"time"
)

var job *cron.Cron

func InitCrons() {

	//在这里启动消费者
	kafkaConsumer.StartNormalConsumer()
	kafkaConsumer.StartDelayConsumer()

	job = cron.New(cron.WithSeconds())

	//每天零点持久化runtime/log文件
	job.AddFunc(consts.CRON_EVERYDAY_MIDNIGHT, StoreRuntimeLogFile)

	//每天零点向所有用户发送日报：新增了多少名粉丝
	job.AddFunc(consts.CRON_EVERYDAY_MIDNIGHT, DailyReport)

	job.Start()
}

// AddTask 添加定时任务
func AddTask(spec string, task func()) error {
	_, err := job.AddFunc(spec, task)
	if err != nil {
		global.Logger.Errorf("添加定时任务出错：%s", err.Error())
		return errors.New("添加定时任务出错：" + err.Error())
	}
	return nil
}

// 备份前一天的日志，同时创建出下一天的日志目录，避免读取不到文件的情况
var (
	files = []string{
		fmt.Sprintf("./runtime/log/%s/error.log", time.Now().Add(-1*time.Hour).Format(time.DateOnly)),
		fmt.Sprintf("./runtime/log/%s/info.log", time.Now().Add(-1*time.Hour).Format(time.DateOnly)),
	}
)

// StoreRuntimeLogFile 持久化日志文件
func StoreRuntimeLogFile() {
	for _, filePath := range files {
		if err := ProcessFile(filePath); err != nil {
			global.Logger.Errorf("持久化日志出错,文件路径为%s", filePath)
		}
	}
}

func ProcessFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File %s does not exist, skipping...\n", path)
			return nil
		}
		log.Fatal("打开文件出错:" + err.Error())
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		entry := new(cron_events.RuntimeLogEntry)
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			log.Fatalln("读取日志行出错" + err.Error())
			continue
		}
		if err := entry.Create(); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}
	return nil
}

// todo：其实可以新增一个日报或者周报，通知用户过去的一周新增了多少粉丝，多少播放量;就是去查对应的信息然后写到notice库里面
//
//	日报应该只对活跃用户(被访问多的用户)生成；周报可以对所有用户生成
func DailyReport() {
	var content string = "您的昨日报告"
	users := new(userModel.UserList)
	err := users.GetAllUserIds()
	if err != nil {
		global.Logger.Errorln("获取所有用户err ：" + err.Error())
		return
	}
	at := new(attention.Attention)

	for _, user := range *users {
		count, err := at.GetNewAddAttentionByTime(time.Now().Add(-24*time.Hour).Format(time.DateTime), user.ID)
		if err != nil {
			global.Logger.Errorln(err.Error())
			return
		}
		content = fmt.Sprintf("%s:昨日新增了%d名粉丝", content, count)
		ne := new(noticeModel.Notice)
		if err = ne.AddNotice(user.ID, 0, 0, noticeModel.DailyReport, content); err != nil {
			global.Logger.Errorln("发送日报err ：" + err.Error())
			return
		}
	}
	global.Logger.Infoln("向所有用户发送日报成功")
}
