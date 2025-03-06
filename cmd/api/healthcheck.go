package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 返回服务器的状态 运行环境 版本
func (app *application) healthcheckHandler(c *gin.Context) {
	c.String(http.StatusOK, "status:available\n")
	c.String(http.StatusOK, "environment:%s\n", app.config.env)
	c.String(http.StatusOK, "version:%s\n", version)
}
