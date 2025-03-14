package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 生成错误日志
func (app *application) logError(c *gin.Context, err error) {
	// 记录当前的访问方法与访问路径
	app.logger.PrintError(err, map[string]string{
		"request_method": c.Request.Method,
		"request_url":    c.Request.URL.String(),
	})
}

// 使用Json向响应体发送错误信息 使用interface{}描述数据可以使Json内容更灵活多变
func (app *application) errorResponse(c *gin.Context, status int, message interface{}) {
	// 先用自定义类型美化输出Json的格式
	env := envelop{"error": message}
	// 向响应体写入数据
	app.writeJson(c, status, env, nil)
}

// 记录服务器在运行时发生的错误(sql查询等非人为造成的错误)
func (app *application) serverErrorResponse(c *gin.Context, err error) {
	// 生成当前错误的日志
	app.logError(c, err)

	msg := "the server encountered a problem and could not process your request"
	// 向响应体发送错误信息
	app.errorResponse(c, http.StatusInternalServerError, msg)
}

// 发送NOT FOUND状态码与Json内容(找不到对应id的记录)
func (app *application) notFoundResponse(c *gin.Context) {
	// 初始化Json字符串
	msg := "the requested resource could not be found"
	// 输出到响应体
	app.errorResponse(c, http.StatusNotFound, msg)
}

// 提示当前请求方法不被允许
func (app *application) methodNotAllowedResponse(c *gin.Context) {
	// 初始化Json字符串
	msg := fmt.Sprintf("the %s method is not supported for this resource", c.Request.Method)
	// 输出到响应体
	app.errorResponse(c, http.StatusMethodNotAllowed, msg)
}

// 返回Bad request信息 人为造成的错误
func (app *application) badRequestResponse(c *gin.Context, err error) {
	app.errorResponse(c, http.StatusBadRequest, err.Error())
}

// 输出验证错误信息
func (app *application) failedValidationResponse(c *gin.Context, errors map[string]string) {
	// 直接将整个用于记录错误的字典以JSON形式输出
	app.errorResponse(c, http.StatusUnprocessableEntity, errors)
}

// 返回修改冲突
func (app *application) editConflictResponse(c *gin.Context) {
	msg := "unable to update the record due to an edit conflict, please try again later"
	// 传入HTTP冲突状态码
	app.errorResponse(c, http.StatusConflict, msg)
}
