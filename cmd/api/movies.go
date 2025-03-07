package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (app *application) createMovieHandler(c *gin.Context) {
	c.String(http.StatusOK, "create a new movie...")
}
func (app *application) showMovieHandler(c *gin.Context) {
	// 使用抽象出来的数据读取模块提取id
	id, err := app.readIDParam(c)
	if err != nil {
		c.String(http.StatusNotFound, "movie id not found\n")
		return
	}
	c.String(http.StatusOK, "show the details of movie:%d\n", id)
}
