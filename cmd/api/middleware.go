package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"greenlight.vdebu.net/internal/data"
	validator2 "greenlight.vdebu.net/internal/validator"
	"net"
	"strings"
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
		if app.config.limiter.enable {
			// 若速率限制是开启的
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
		}
		// 调用下一个中间件
		context.Next()
	}
}

// 通过从r.context中提取相关内容判断用户时候有认证权限
func (app *application) authenticate() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 告诉浏览器根据Authorization的值进行缓存 -> Vary: Authorization
		context.Header("Vary", "Authorization")
		// 尝试从表头提取Authorization字段的信息
		authorizationHeader := context.GetHeader("Authorization")
		// 如果是空的则将当前用户设置为匿名用户
		if authorizationHeader == "" {
			app.contextSetUser(context.Request, data.AnonymousUser)
			// 直接调用下一个中间件不执行后续的代码
			context.Next()
			// 必须直接return!!!
			// 简单的调用中间件链并不会终止当前中间件代码的后续运行
			return
		}
		// 存储在表头中的结构应该是:Authorization: Bearer <Token>
		// 提取成功后尝试进行切分并检查是否如预期
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(context)
			return
		}
		// 提取Token进行有效性检测
		token := headerParts[1]
		v := validator2.New()
		// 若有效性验证失败返回无效的认证秘钥而不是通常使用的无效的验证
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(context)
			return
		}
		// 尝试提取当前秘钥的相关用户
		user, err := app.models.User.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			// 判断是否是查无此人
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				// 找不到相关的用户说明认证信息是无效的
				app.invalidAuthenticationTokenResponse(context)
			default:
				// 若为其他错误则是服务器内部发生的
				app.serverErrorResponse(context, err)
			}
			return
		}
		// 验证成功更新当前请求的context信息
		context.Request = app.contextSetUser(context.Request, user)
		// 调用下一个中间件
		context.Next()
	}
}
