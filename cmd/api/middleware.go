package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"net"
	"sync"
	"time"
)

// 检测Panic
func (app *application) recoverPanic() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 在发生了panic后就会被执行
		defer func() {
			if err := recover(); err != nil {
				// 设置连接关闭 go的httpserver会自动关闭当前连接
				context.Header("Connection", "close")
				// recover返回的是interface{}使用fmt.Errorf将其格式化
				app.serverErrorResponse(context, fmt.Errorf("%s", err))
			}
		}()
		// 调用下一个中间件
		context.Next()
	}
}

// 使用令牌桶限制访问速率
func (app *application) rateLimiter() gin.HandlerFunc {
	// 使用结构体存储某个客户端上次访问的时间与该客户端对应的限速器
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time // 记录下上次访问API的时间
	}
	var (
		mu      sync.Mutex                 // 创建锁
		clients = make(map[string]*client) // 使用map对每一个IP的令牌桶进行映射 id -> limiter & lastSeen
	)
	// 启动goroutine用于清理已经失效的ip
	go func() {
		for {
			// 每间隔一分钟进行一次清理
			time.Sleep(time.Minute)
			// 对即将执行的操作上锁
			mu.Lock()
			for ip, client := range clients {
				// 对访问间隔超过三分钟的ip进行清理
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// 解锁
			mu.Unlock()
		}
	}()
	return func(context *gin.Context) {
		// 提取每一个请求的ip
		ip, _, err := net.SplitHostPort(context.Request.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(context, err)
			return
		}
		// 上锁 准备对clients进行操作
		mu.Lock()
		// 检查当前ip是否已存在 如果不存在就进行初始化
		if _, found := clients[ip]; !found {
			// 初始化 ip -> client
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}
		// 记录下当前访问的时间
		clients[ip].lastSeen = time.Now()
		// 针对每一个IP使用Allow会消耗令牌 如果没有令牌会返回False
		if !clients[ip].limiter.Allow() {
			// 避免死锁
			mu.Unlock()
			// 返回请求繁忙
			app.rateLimitExceededResponse(context)
			return
		}
		// 操作完成解锁
		mu.Unlock()
		// 调用下一个中间件
		context.Next()
	}
}
