package main

import (
	"errors"
	"expvar"
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/gin-gonic/gin"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"greenlight.vdebu.net/internal/data"
	validator2 "greenlight.vdebu.net/internal/validator"
	"net/http"
	"strconv"
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
			ip := realip.FromRequest(context.Request)
			// 上锁 准备对clients进行操作
			mu.Lock()
			// 检查当前ip是否已存在 如果不存在就进行初始化
			if _, found := clients[ip]; !found {
				// 初始化 ip -> client
				// 使用配置文件中存储的数据对令牌桶进行初始化
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
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
		//app.logger.PrintInfo("authenticate middleware called", nil)
		// 告诉浏览器根据Authorization的值进行缓存 -> Vary: Authorization
		context.Writer.Header().Add("Vary", "Authorization")
		// 尝试从表头提取Authorization字段的信息
		authorizationHeader := context.GetHeader("Authorization")
		// 如果是空的则将当前用户设置为匿名用户
		if authorizationHeader == "" {
			app.contextSetUser(context, data.AnonymousUser)
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
		app.contextSetUser(context, user)
		// 调用下一个中间件
		context.Next()
	}
}

// 检查当前是否是匿名用户(未认证)
func (app *application) requireAuthenticatedUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		user := app.contextGetUser(context)
		if user.IsAnonymous() {
			app.authenticationRequireResponse(context)
			return
		}
		// 调用下一个中间件
		context.Next()
	}
}

// 检查用户是否已认证并且账号已激活(应用在权限敏感的网页)
func (app *application) requireActivatedUser() gin.HandlerFunc {
	return func(context *gin.Context) {
		user := app.contextGetUser(context)
		// 不是匿名用户判断当前账号是否激活
		if !user.Activated {
			app.inactivatedAccountResponse(context)
			return
		}
		// 调用下一个中间件
		context.Next()
	}
}

// 接收權限的代碼映射數據庫中的權限類型
func (app *application) requirePermission(code string) gin.HandlerFunc {
	return func(context *gin.Context) {
		// 先從ctx中提取出當前用戶的信息
		user := app.contextGetUser(context)
		// 從數據庫中查詢當前用戶擁有的權限
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(context, err)
			return
		}
		// 檢查提供的權限代碼是否在查詢到的列表中
		if !permissions.Include(code) {
			// 寫入相應響應體
			app.notPermittedResponse(context)
			return
		}
		// 有對應的權限則調用下一個中間件
		context.Next()
	}
}

// 設置允許跨源請求的請求頭
func (app *application) enableCORS() gin.HandlerFunc {
	return func(context *gin.Context) {
		// 確保服務器不會將不同用戶的私有相應緩存並共享
		// 注意這裡使用Add方法對表頭進行操作防止先前參數被覆蓋
		context.Writer.Header().Add("Vary", "Origin")
		// 添加表示动态变化的请求头
		context.Writer.Header().Add("Vary", "Access-Control-Request-Method")
		// Vary: Origin, Access-Control-Request-Method
		// 從響應頭提取訪問源
		origin := context.GetHeader("Origin")
		// 如果origin不為空並且設置的信任的請求源再設置相應的許可
		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			// 判斷當前請求源是否處於信任列表裡
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					// 處於列表中澤設置許可
					context.Header("Access-Control-Allow-Origin", origin)
					// 查看当前传入的请求是否是预检请求(OPTIONS)与CORS相关表头是否存在
					if context.Request.Method == http.MethodOptions && context.GetHeader("Access-Control-Request-Method") != "" {
						// 写入允许的请求方法与请求头
						context.Header("Access-Control-Allow-Methods", "OPTIONS, POST, PUT, PATCH, DELETE")
						context.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
						// 返回正确状态码并提前结束预检请求
						// 同样也需要调用Abort
						context.AbortWithStatus(http.StatusNoContent)
						return
					}
				}
			}
		}
		// 調用下一個中間件
		context.Next()
	}
}

// 记录收到的请求数与发送的请求数及完成请求所耗时间
func (app *application) metrics() gin.HandlerFunc {
	// 初始化在监控节点中要展示的信息
	totalRequestReceived := expvar.NewInt("total_requests_received")
	totalRequestSent := expvar.NewInt("total_response_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_us")
	// 使用三方库记录收到的HTTP code的种类与数目
	totalRequestSentByStatus := expvar.NewMap("total_response_sent_by_status")
	return func(context *gin.Context) {
		// 收到的请求数自增
		totalRequestReceived.Add(1)
		// 调用下一个中间件
		// 将控制权传递给下一个中间件或者最终的路由处理器
		// context.Next() 是阻塞调用，会顺序执行链中后续的所有中间件和最终的处理器。当这些处理器全部执行完毕后，控制权才会返回到当前中间件中继续执行下面的代码。
		//也就是说，下半部分代码在响应生成完毕后才会执行。
		//context.Next()

		// 使用三方库捕获响应体代码请求持续时间与成功写入响应体的字节
		metrics := httpsnoop.CaptureMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 对Gin的next函数进行封装
			context.Next()
		}), context.Writer, context.Request)
		// 当返回中间件链的时候标记为请求已发送
		totalRequestSent.Add(1)
		// 写入完成耗时
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		// 记录下当前请求的代码(将string转换成int)
		totalRequestSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	}
}
