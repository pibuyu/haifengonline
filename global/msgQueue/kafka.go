package msgQueue

import (
	"context"
	"haifengonline/global/config"

	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
)

var msgConfig = config.Config.KafkaConfig

type KafkaProducerPool struct {
	ConnPool chan *kafka.Conn
}

// ReturnsInstance 不要这样写，直接像测试类一样，构造一个kafka连接并且返回就行
func ReturnsNormalInstance() *kafka.Conn {

	conn, err := GetNormalProducerConn()
	if err != nil {
		log.Fatalf("获取kafka连接出错：%v", err)
	}
	return conn
}

func ReturnsDelayInstance() *kafka.Conn {
	conn, err := GetDelayProducerConn()
	if err != nil {
		log.Fatalf("获取kafka连接出错：%v", err)
	}
	return conn
}

func GetNormalProducerConn() (*kafka.Conn, error) {
	//conn, err := kafka.DialLeader(context.Background(), "tcp", consts.KafkaServerAddr, consts.KafkaTopic, 0)
	conn, err := kafka.DialLeader(context.Background(), "tcp", msgConfig.Server, msgConfig.NormalTopic, 0)
	if err != nil {
		return nil, fmt.Errorf("创建kafka连接出错：%v", err)
	}
	return conn, nil
}

func GetDelayProducerConn() (*kafka.Conn, error) {
	//conn, err := kafka.DialLeader(context.Background(), "tcp", consts.KafkaServerAddr, consts.DelayQueue, 0)
	conn, err := kafka.DialLeader(context.Background(), "tcp", msgConfig.Server, msgConfig.DelayTopic, 0)
	if err != nil {
		return nil, fmt.Errorf("创建kafka连接出错：%v", err)
	}
	return conn, nil
}
