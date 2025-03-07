package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 生成错误日志
func (app *application) logError(c *gin.Context, err error) {
	app.logger.Println(err)
}

// 使用Json向响应体发送错误信息 使用interface{}描述数据可以使Json内容更灵活多变
func (app *application) errorResponse(c *gin.Context, status int, message interface{}) {
	// 先用自定义类型美化输出Json的格式
	env := envelop{"error": message}
	// 向响应体写入数据
	app.writeJson(c, status, env, nil)
}

// 记录服务器在运行时发生的错误
func (app *application) serverErrorResponse(c *gin.Context, err error) {
	// 生成当前错误的日志
	app.logError(c, err)

	msg := "the server encountered a problem and could not process your request..."
	// 向响应体发送错误信息
	app.errorResponse(c, http.StatusInternalServerError, msg)
}

// 发送NOT FOUND状态码与Json内容
func (app *application) notFoundResponse(c *gin.Context) {
	// 初始化Json字符串
	msg := "the requested resource could not be found..."
	// 输出到响应体
	app.errorResponse(c, http.StatusNotFound, msg)
}

func (app *application) methodNotAllowedResponse(c *gin.Context) {
	// 初始化Json字符串
	msg := fmt.Sprintf("the %s method is not supported for this resource...", c.Request.Method)
	// 输出到响应体
	app.errorResponse(c, http.StatusMethodNotAllowed, msg)
}
