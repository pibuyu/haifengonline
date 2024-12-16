package video

import (
	"context"
	"haifengonline/global"
	"haifengonline/logic/contribution"
	"haifengonline/models/common"
	"haifengonline/models/home/rotograph"
	quartzImpl "haifengonline/quartzImpl"

	"fmt"
	"github.com/reugn/go-quartz/quartz"
	"time"
)

// 不需要参数
type UpdateHomeRotograph struct {
}

//type Rotograph struct {
//	common.PublicModel
//	Title string         `json:"title" gorm:"column:title"`
//	Cover datatypes.JSON `json:"cover" gorm:"column:cover"`
//	Color string         `json:"color" gorm:"column:color" `
//	Type  string         `json:"type" gorm:"column:type"`
//	ToId  uint           `json:"to_id" gorm:"column:to_id"`
//}

// 每天定时更新热门轮播图
func (this *UpdateHomeRotograph) Execute(ctx context.Context) error {
	//todo:找到最热门的两个视频和一个最热门的专栏，放进rotograph表中去
	heatestVideos, err := contribution.GetTop2HeatVideos()
	if err != nil {
		return err
	}
	heatestArticle, err := contribution.GetHeatestArticle()
	if err != nil {
		return err
	}
	//将上述结果插入转为rotograph.Rotograph类型，并合并为列表
	var rotographs []rotograph.Rotograph
	for _, video := range heatestVideos {
		rotographs = append(rotographs, rotograph.Rotograph{
			PublicModel: common.PublicModel{},
			Title:       video.Title,
			Cover:       video.Cover,
			Color:       "rgb(116,82,81)",
			Type:        "video",
			ToId:        video.ID,
		})
	}

	rotographs = append(rotographs, rotograph.Rotograph{
		PublicModel: common.PublicModel{},
		Title:       heatestArticle.Title,
		Cover:       heatestArticle.Cover,
		Color:       "rgb(116,82,81)",
		Type:        "article",
		ToId:        heatestArticle.ID,
	})

	//开启一个事务插入上面的三条记录
	tx := global.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for _, r := range rotographs {
		if err := r.Create(); err != nil {
			tx.Rollback()
			return err
		}

	}
	return tx.Commit().Error
}

// 重写Description函数
func (this *UpdateHomeRotograph) Description() string {
	return "每天24:00定时更新轮播图的方法"
}

func updateRotograph(ctx context.Context) error {
	scheduler := quartzImpl.Pool.GetScheduler()
	defer quartzImpl.Pool.ReleaseScheduler(scheduler)
	scheduler.Start(ctx)

	job := &UpdateHomeRotograph{}
	detail := quartz.NewJobDetail(job, quartz.NewJobKey(fmt.Sprintf("updateRotograph_%s", time.Now().Format("2006-01-02 15:04:05"))))

	cronTrigger, err := quartz.NewCronTriggerWithLoc("0 0 0 * * ?", time.Local)
	if err != nil {
		return fmt.Errorf("创建cronTrigger failed:%v", err)
	}
	err = scheduler.ScheduleJob(detail, cronTrigger)

	if err != nil {
		return fmt.Errorf("定时更新rotograph任务设置失败:%v", err)
	}

	global.Logger.Infof("定时任务设置成功，每天00:00清空 lv_home_rotograph 表")
	return nil
}
