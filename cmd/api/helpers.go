package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

// 用于折嵌套Json 使用key匹配interface{}数据
type envelop map[string]interface{}

// 将用于提取id的代码逻辑抽象成一个模块
func (app *application) readIDParam(c *gin.Context) (int64, error) {
	params := c.Param("id")
	// 提取数字并按照制定格式转化
	id, err := strconv.ParseInt(params, 10, 64)
	if err != nil || id < 1 {
		return 0, err
	}
	return id, nil
}

// 发送json数据到响应体 将data的类型更改为自定义类型用于折叠Json
func (app *application) writeJson(c *gin.Context, status int, data envelop, header http.Header) {
	// 使json的输出更加美观
	c.IndentedJSON(status, data)
	// 向响应体中添加传入的表头
	for key, val := range header {
		// 将 []string 转换为 string，多个值以逗号分隔
		c.Header(key, strings.Join(val, ","))
	}
	//app.logger.Println("we are here...")
}
