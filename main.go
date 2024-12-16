package main

import (
	"haifengonline/crons"
	_ "haifengonline/global/database/mysql" //初始化mysql
	_ "haifengonline/global/database/redis" //初始化redis
	"haifengonline/router"

	_ "haifengonline/utils/socket"  //初始化socket
	_ "haifengonline/utils/testing" //运行环境检查
)

func main() {
	//这里还包括了消费者函数的启动
	crons.InitCrons()
	//注册路由
	router.InitRouter()
}
