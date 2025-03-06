package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (app *application) createMovieHandler(c *gin.Context) {
	c.String(http.StatusOK, "create a new movie...")
}
func (app *application) showMovieHandler(c *gin.Context) {
	// 直接提取url中的参数
	params := c.Param("id")
	// 将提取到的字符串进行格式转换
	id, err := strconv.Atoi(params)
	if err != nil {
		c.String(http.StatusNotFound, "movie id:%v not found\n", params)
		return
	}
	c.String(http.StatusOK, "show the details of movie:%d\n", id)
}
