package main

import (
	"crawlab/config"
	"crawlab/database"
	"crawlab/middlewares"
	"crawlab/routes"
	"crawlab/services"
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"runtime/debug"
)

func main() {
	app := gin.Default()

	// 初始化配置
	if err := config.InitConfig(""); err != nil {
		panic(err)
	}
	log.Info("初始化配置成功")

	// 初始化日志设置
	logLevel := viper.GetString("log.level")
	if logLevel != "" {
		log.SetLevelFromString(logLevel)
	}
	log.Info("初始化日志设置成功")

	// 初始化Mongodb数据库
	if err := database.InitMongo(); err != nil {
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化Mongodb数据库成功")

	// 初始化Redis数据库
	if err := database.InitRedis(); err != nil {
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化Redis数据库成功")

	if services.IsMaster() {
		// 初始化定时任务
		if err := services.InitScheduler(); err != nil {
			debug.PrintStack()
			panic(err)
		}
		log.Info("初始化定时任务成功")
	}

	// 初始化任务执行器
	if err := services.InitTaskExecutor(); err != nil {
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化任务执行器成功")

	// 初始化节点服务
	if err := services.InitNodeService(); err != nil {
		panic(err)
	}
	log.Info("初始化节点配置成功")

	// 初始化爬虫服务
	if err := services.InitSpiderService(); err != nil {
		debug.PrintStack()
		panic(err)
	}
	log.Info("初始化爬虫服务成功")

	// 以下为主节点服务
	if services.IsMaster() {
		// 中间件
		app.Use(middlewares.CORSMiddleware())

		// 路由
		// 节点
		app.GET("/nodes", routes.GetNodeList)               // 节点列表
		app.GET("/nodes/:id", routes.GetNode)               // 节点详情
		app.POST("/nodes/:id", routes.PostNode)             // 修改节点
		app.GET("/nodes/:id/tasks", routes.GetNodeTaskList) // 节点任务列表
		app.GET("/nodes/:id/system", routes.GetSystemInfo)  // 节点任务列表
		// 爬虫
		app.GET("/spiders", routes.GetSpiderList)              // 爬虫列表
		app.GET("/spiders/:id", routes.GetSpider)              // 爬虫详情
		app.PUT("/spiders", routes.PutSpider)                  // 上传爬虫
		app.POST("/spiders", routes.PublishAllSpiders)         // 发布所有爬虫
		app.POST("/spiders/:id", routes.PostSpider)            // 修改爬虫
		app.POST("/spiders/:id/publish", routes.PublishSpider) // 发布爬虫
		app.DELETE("/spiders/:id", routes.DeleteSpider)        // 删除爬虫
		app.GET("/spiders/:id/tasks", routes.GetSpiderTasks)   // 爬虫任务列表
		// 任务
		app.GET("/tasks", routes.GetTaskList)                                 // 任务列表
		app.GET("/tasks/:id", routes.GetTask)                                 // 任务详情
		app.PUT("/tasks", routes.PutTask)                                     // 派发任务
		app.DELETE("/tasks/:id", routes.DeleteTask)                           // 删除任务
		app.POST("/tasks/:id/cancel", routes.CancelTask)                      // 取消任务
		app.GET("/tasks/:id/log", routes.GetTaskLog)                          // 任务日志
		app.GET("/tasks/:id/results", routes.GetTaskResults)                  // 任务结果
		app.GET("/tasks/:id/results/download", routes.DownloadTaskResultsCsv) // 下载任务结果
		// 定时任务
		app.GET("/schedules", routes.GetScheduleList)       // 定时任务列表
		app.GET("/schedules/:id", routes.GetSchedule)       // 定时任务详情
		app.PUT("/schedules", routes.PutSchedule)           // 创建定时任务
		app.POST("/schedules/:id", routes.PostSchedule)     // 修改定时任务
		app.DELETE("/schedules/:id", routes.DeleteSchedule) // 删除定时任务
	}

	// 路由ping
	app.GET("/ping", routes.Ping)

	// 运行服务器
	host := viper.GetString("server.host")
	port := viper.GetString("server.port")
	if err := app.Run(host + ":" + port); err != nil {
		panic(err)
	}
}
