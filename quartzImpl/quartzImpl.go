package quartz

import (
	"github.com/reugn/go-quartz/quartz"
	"haifengonline/consts"
)

var Pool = InitQuartzPool(consts.QuartzPoolSize)

func InitQuartzPool(size int) *SchedulerPool {
	return NewSchedulerPool(size)
}

type SchedulerPool struct {
	pool chan quartz.Scheduler
}

// NewSchedulerPool 初始化连接池
func NewSchedulerPool(maxSize int) *SchedulerPool {
	pool := make(chan quartz.Scheduler, maxSize)
	for i := 0; i < maxSize; i++ {
		scheduler := quartz.NewStdScheduler()
		pool <- scheduler
	}
	return &SchedulerPool{pool: pool}
}

// GetScheduler 从连接池获取一个调度器
func (p *SchedulerPool) GetScheduler() quartz.Scheduler {
	return <-p.pool
}

// ReleaseScheduler 调度器工作完成释放回连接池
func (p *SchedulerPool) ReleaseScheduler(schd quartz.Scheduler) {
	p.pool <- schd
}
