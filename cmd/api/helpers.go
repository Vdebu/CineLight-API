package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
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
