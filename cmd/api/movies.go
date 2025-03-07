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
		app.notFoundResponse(c)
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
	// 将电影的信息以json的形式输出 使用自定义类型进行封装以呈现出嵌套展示的效果
	app.writeJson(c, http.StatusOK, envelop{"movie": movie}, nil)
}
