package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 返回服务器的状态 运行环境 版本
func (app *application) healthcheckHandler(c *gin.Context) {
	// 定义raw json
	//js := `{"status":"available","environment":%q,"version":%q}`
	// 初始化json
	// 也可以通过gin.H{"status":"available","environment":app.config.env,"version":version}进行定义
	//js = fmt.Sprintf(js, app.config.env, version)

	// 设置表头参数告诉客户端输送的是json数据
	//c.Header("Content-Type", "application/json")
	// 像响应体输出json数据 会自动设置表头参数 Content-Type as "application/ json".
	// c.String(http.StatusOK, js) 在这里用标准string输出浏览器才会自动检测为json格式 而不能用c.Json输出

	// c.Json同样可以解析map[string]string -- map[string]string
	//data := map[string]string{
	//	"status":      "available",
	//	"environment": app.config.env,
	//	"version":     version,
	//}
	//data := make(chan string)
	data := gin.H{"status": "available", "environment": app.config.env, "version": version}
	// 自动初始化并输出json数据
	app.writeJson(c, http.StatusOK, data, nil)
}
