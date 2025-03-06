package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// 定义当前版本信息 后续会使用自动生成的手段进行改进
const version = " 1.0.0"

// 存储服务器的配置信息
type config struct {
	// 端口
	port int
	// 运行环境
	env string
}

// 注入依赖
type application struct {
	config config
	logger *log.Logger
}

func main() {
	// 添加命令行初始化服务器参数的方法
	var cfg config
	// 服务器监听的端口 默认3939
	flag.IntVar(&cfg.port, "port", 3939, "API server port")
	// 服务器的环境信息
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")
	// 解析命令行参数
	flag.Parse()
	// 初始化服务器内部的日志工具
	// 输出时间与日期
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	// 初始化依赖
	app := &application{
		config: cfg,
		logger: logger,
	}
	// 初始化服务器信息
	srv := http.Server{
		// 初始化端口
		Addr: fmt.Sprintf(":%d", cfg.port),
		// 初始化路由
		Handler: app.routers(),
		// 初始化各种操作的超时时间
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	// 启动服务器
	logger.Printf("starting %s server on %s\n", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
