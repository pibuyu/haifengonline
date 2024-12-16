package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sethvargo/go-retry"
	"haifengonline/global/config"
	"log"
	"time"
)

var Db *redis.Client

func ReturnsInstance() *redis.Client {
	var err error
	var redisConfig = config.Config.RConfig
	b := retry.NewFibonacci(10 * time.Second) //重试的斐波那契机制，最大重试间隔时间为10秒
	ctx := context.Background()
	if err := retry.Do(ctx, retry.WithMaxRetries(5, b), func(ctx context.Context) error {
		Db = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%d", redisConfig.IP, redisConfig.Port),
			Password:     redisConfig.Password, // no password set
			DB:           0,                    // use default DB
			PoolSize:     20,                   //最大连接数，默认为4*cpu个数
			MinIdleConns: 5,                    //最少活跃连接数
		})
		_, err = Db.Ping().Result()
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		// 多次重连仍旧失败
		//这个地方不能用logger打印，不然循环依赖报错
		//global.Logger.Errorf("重连5次后redis连接失败，可能是redis客户端未启动- %v", err)
		log.Fatalf("重试5次后仍然无法连接redis，请排查redis服务端是否启动/配置信息是否正确，错误信息为： %v \n", err)
	}
	return Db
}
