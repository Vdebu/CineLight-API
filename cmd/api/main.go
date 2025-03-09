package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq" //隐式导入sql驱动
)

// 定义当前版本信息 后续会使用自动生成的手段进行改进
const version = " 1.0.0"

// 存储服务器的配置信息
type config struct {
	port int    // 端口
	env  string // 运行环境
	db   struct {
		dsn          string // 在服务器配置中存储dsn
		maxOpenConns int    // 最大同时建立的连接数(active + idle)
		maxIdleConns int    // 最大惰性连接数 maxIdleConns <= maxOpenConns
		maxIdleTime  string // 在连接持续处于惰性一段时间后将其关闭
	}
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
	// 默认从系统的环境变量中获取服务器的数据库DSN(data source name)
	// PostgreSQL驱动可能会使用 SSL连接如果服务器没有启用 SSL需要在 DSN 中添加参数来禁用SSL
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	// 服务器数据库连接池的配置
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// 解析命令行参数
	flag.Parse()
	// 初始化服务器内部的日志工具
	// 输出时间与日期
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	//logger.Println("dsn:", cfg.db.dsn)
	// 初始化数据库链接
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	// 程序结束时关闭数据库的连接
	defer db.Close()
	logger.Println("db connection established...")
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
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// 尝试连接数据库 返回数据库连接池sql.DB
func openDB(cfg config) (*sql.DB, error) {
	// 传入数据库信息
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// 设置建立的最大连接数
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	// 设置最大的惰性链接数
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	// 设置清理惰性链接的时间
	// 使用ParseDuration对字符串形式的时间进行转换(15m)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)
	// 创建Context 在5秒连接超时后关闭连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 超时后关闭连接
	defer cancel()

	// 检查连接是否建立
	// 如果在5秒内没有ping成功就会返回错误
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
