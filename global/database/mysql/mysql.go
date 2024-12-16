package mysql

import (
	"context"
	"haifengonline/global/config"

	"fmt"
	"github.com/sethvargo/go-retry"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	globalLog "haifengonline/global/logrus"
	"log"
	"os"
	"time"
)

var Db *gorm.DB

type MyWriter struct {
	log *logrus.Logger
}

type MysqlDB struct {
	DB *gorm.DB
}

// Printf 实现gorm/logger.Writer接口
func (m *MyWriter) Printf(format string, v ...interface{}) {
	m.log.Errorf(fmt.Sprintf(format, v...))
}

func NewMyWriter() *MyWriter {
	instance := globalLog.ReturnsInstance()
	return &MyWriter{log: instance}
}

func ReturnsInstance() *gorm.DB {
	var mysqlConfig = config.Config.SqlConfig
	//sql日志记录
	myLogger := logger.New(
		//设置Logger
		//NewMyWriter(),

		//输出在控制台，方便debug
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel:                  logger.Silent, //仅仅在控制台输出指定Debug的语句
			IgnoreRecordNotFoundError: true,          // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,          // 禁用彩色打印
		},
	)
	b := retry.NewFibonacci(10 * time.Second)
	ctx := context.Background()
	if err := retry.Do(ctx, retry.WithMaxRetries(5, b), func(ctx context.Context) error {
		// 创建链接
		var err error
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local", mysqlConfig.User, mysqlConfig.Password, mysqlConfig.IP, mysqlConfig.Port, mysqlConfig.Database)
		Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: myLogger,
		})
		if err != nil {
			return err
		}
		if Db.Error != nil {
			return err
		}
		return nil
	}); err != nil {
		// handle error
		log.Fatalf("重试5次后仍然无法连接mysql，请排查mysql服务端是否启动/配置信息是否正确，错误信息为： %v \n", err)
	}
	return Db
}
