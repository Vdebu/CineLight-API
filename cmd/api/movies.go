package main

import (
	"github.com/gin-gonic/gin"
	"greenlight.vdebu.net/internal/data"
	"net/http"
	"time"
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
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Project Sekai",
		Year:      2025,
		Runtime:   3939,
		Genres:    []string{"anime", "miku"},
		Version:   1,
	}
	// 将电影的信息以json的形式输出
	app.writeJson(c, http.StatusOK, movie, nil)
}
