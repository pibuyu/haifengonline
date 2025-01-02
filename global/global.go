package global

import (
	"github.com/go-redis/redis"
	sensitive "github.com/pibuyu/sensitive_words_filter"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"haifengonline/Init/sensitiveWordsFilter"
	"haifengonline/global/config"
	"haifengonline/global/database/mysql"
	RedisDbFun "haifengonline/global/database/redis"
	log "haifengonline/global/logrus"
)

// 在这里执行那些实例化的init函数，然后返回预定义的对象呗
func init() {
	Logger = log.ReturnsInstance()
	RedisDb = RedisDbFun.ReturnsInstance()
	Db = mysql.ReturnsInstance()
	Config = config.ReturnsInstance()
	//普通队列的生产者
	//NormalProducer = msgQueue.ReturnsNormalInstance()
	//延迟队列的生产者
	//DelayProducer = msgQueue.ReturnsDelayInstance()
	//构造一个敏感词过滤器对象
	Filter = sensitiveWordsFilter.InitFilter()
}

var (
	Logger         *logrus.Logger
	Config         *config.Info
	Db             *gorm.DB
	RedisDb        *redis.Client
	NormalProducer *kafka.Conn
	DelayProducer  *kafka.Conn
	Filter         *sensitive.Manager
)
