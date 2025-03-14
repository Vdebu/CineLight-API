package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
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
	// 每秒新增两个令牌 最多持有四个令牌
	limiter := rate.NewLimiter(2, 4)
	return func(context *gin.Context) {
		// 使用Allow会消耗令牌 如果没有令牌会返回False
		if !limiter.Allow() {
			app.rateLimitExceededResponse(context)
			return
		}
		context.Next()
	}
}
