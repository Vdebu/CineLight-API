package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

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

// 发送json数据到响应体
func (app *application) writeJson(c *gin.Context, status int, data interface{}, header http.Header) {
	// 数据初始化问题 不能确定传入的数据是否符合json格式 gin也没有提供内置函数进行判断
	// gin的c.Json也是调用json.Marshal但是不会自动检查并报告错误
	// 进行实际操作前进行人为调用json.Marshal并进行错误检测
	//if _, err := json.Marshal(data); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid json data"})
	//	return
	//}
	// 确定是能转换成json格式后 将提供的数据转换成json并输出到响应体中

	// gin在进行json格式转换时是会判断格式并返回错误的 如果格式不正确会抛出错误并停止执行后续的所有代码
	c.JSON(status, data)
	// 向响应体中添加传入的表头
	for key, val := range header {
		// 将 []string 转换为 string，多个值以逗号分隔
		c.Header(key, strings.Join(val, ","))
	}
	//app.logger.Println("we are here...")
}
