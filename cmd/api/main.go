package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"greenlight.vdebu.net/internal/data"
	"greenlight.vdebu.net/internal/jsonlog"
	"greenlight.vdebu.net/internal/mailer"
	"os"
	"runtime"
	"strings"
	"sync"
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
	limiter struct {
		rps    float64 // request-per-second 每秒填充的令牌数
		burst  int     // 默认令牌值
		enable bool    // 是否开启速率限制
	}
	smtp struct {
		host     string
		port     int
		username string // 用于发送邮箱的账号
		password string // 用于发送邮箱的账号
		sender   string // 发件人
	}
	cors struct {
		trustedOrigins []string // 受信的跨院網站
	}
}

// 注入依赖
type application struct {
	config config          // 服务器默认配置
	logger *jsonlog.Logger // JSON形式的logger
	models data.Models     // 数据库中的数据模型
	mailer mailer.Mailer   // 邮箱服务
	wg     sync.WaitGroup  // 同步goroutine工作进度 默认0值后续无需进行初始化
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
	// 在makefile中导入环境变量
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	// 服务器数据库连接池的配置
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	// 服务器速率限制的配置
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enable, "limiter-enabled", true, "Enable rate limiter")
	// 邮箱服务器的配置
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "6d1f560db0b87a", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "ca65fbfdf5d908", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@Greenlight.vdebu.net>", "SMTP sender")
	// 受信的跨源請求
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(s string) error {
		// 將輸入的字符串
		cfg.cors.trustedOrigins = strings.Fields(s)
		return nil
	})
	// 解析命令行参数
	flag.Parse()
	// 初始化服务器内部的日志工具
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	//logger.Println("dsn:", cfg.db.dsn)
	// 初始化数据库链接
	db, err := openDB(cfg)
	if err != nil {
		// 使用PrintFatal输出错误信息并结束程序运行
		logger.PrintFatal(err, nil)
	}
	// 程序结束时关闭数据库的连接
	defer db.Close()
	logger.PrintInfo("db connection established...", nil)
	// 初始化模型依赖
	models := data.NewModels(db)
	// 初始化显示服务器版本
	expvar.NewString("version").Set(version)
	// 初始化显示程序goroutine的状态(返回的结果必须要能编码成JSON否则在显示时会被忽略)
	expvar.Publish("goroutine", expvar.Func(func() any {
		// 返回启动的goroutine数量
		return runtime.NumGoroutine()
	}))
	// 初始化显示数据库的状态
	expvar.Publish("database", expvar.Func(func() any {
		// 返回数据库的状态
		return db.Stats()
	}))
	// 初始化显示当前时间
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Local()
	}))
	// 初始化依赖
	app := &application{
		config: cfg,    // 载入服务器配置
		logger: logger, // 初始化默认标准输出，信息为Info的Logger
		models: models, // 嵌入数据模型
		mailer: mailer.New( // 初始化邮件系统
			cfg.smtp.host,
			cfg.smtp.port,
			cfg.smtp.username,
			cfg.smtp.password,
			cfg.smtp.sender,
		),
	}
	// 初始化服务器信息
	err = app.server()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
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
