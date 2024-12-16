package video

import (
	"context"
	"fmt"
	"github.com/reugn/go-quartz/quartz"
	"github.com/segmentio/kafka-go"
	"haifengonline/global"
	"haifengonline/logic/contribution"
	quartzImpl "haifengonline/quartzImpl"
	"time"
)

type PublishVideoJob struct {
	ID uint //要定时发布的视频的id
}

// 定时任务只管将任务加入计时器队列中，具体执行靠的是Execute中的逻辑
func (job *PublishVideoJob) Execute(ctx context.Context) error {
	//global.Logger.Infof("调用execute函数，将id为%d的视频定时发布", job.ID)
	err := contribution.SetIsVisibleById(job.ID)
	if err != nil {
		return err
	}
	return nil
}

func (job *PublishVideoJob) Description() string {
	return "定时发布视频方法"
}

// PublishVideo 根据指定时间发布视频：在指定时间将视频的is_visible字段修改为1

func PublishVideo(waitTime time.Duration, id uint, ctx context.Context) error {

	scheduler := quartzImpl.Pool.GetScheduler()
	defer quartzImpl.Pool.ReleaseScheduler(scheduler)
	scheduler.Start(ctx)

	job := &PublishVideoJob{ID: id}
	detail := quartz.NewJobDetail(job, quartz.NewJobKey(fmt.Sprintf("publishVideo_%d", id)))

	//传递过来的waitTime以纳秒为单位，需要将其转换为以秒为单位然后传给NewRunOnceTrigger作为参数
	waitSeconds := time.Duration(int(waitTime.Seconds())) * time.Second
	err := scheduler.ScheduleJob(detail, quartz.NewRunOnceTrigger(waitSeconds))
	global.Logger.Infof("定时发布视频任务设置成功，将在%v秒后发布视频", waitSeconds.Seconds())
	if err != nil {
		return err
	}
	return nil
}

// todo:将预定发布时间写入延时队列里，再由演示队列的消费者去处理;为什么1次向延时队列提交了两条延时消息？？？？？而且其中的一条被立马转到了及时队列中去
func PublishVideoOnSchedule(shceduleTime time.Time, id uint) error {
	msg := kafka.Message{Value: []byte(fmt.Sprintf("publishVideo_%d", id)), Time: shceduleTime}
	_, err := global.DelayProducer.WriteMessages(msg)
	global.Logger.Infof("定时任务写入延时队列成功:%v", msg)
	if err != nil {
		global.Logger.Errorf("定时发布视频任务写入消息队列失败：%v", err)
		return err
	}
	return nil
}
