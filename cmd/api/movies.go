package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"greenlight.vdebu.net/internal/data"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(c *gin.Context) {
	// 创建结构体存储待解析的JSON数据 这个结构体是完整Movie结构体的子集
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}
	// 使用Gin的内置方法对请求体中的JSON数据进行解析
	if err := c.ShouldBindJSON(&input); err != nil {
		// 向响应体中输出错误信息
		app.errorResponse(c, http.StatusBadRequest, err.Error())
	}
	// 使用Windows发送 必须将引号转义
	// curl -i -d "{\"title\":\"Moana\",\"year\":2016,\"runtime\":107,\"genres\":[\"animation\",\"adventure\"]}" http://localhost:3939/v1/movies
	// 将提取到的信息作为响应体输出
	fmt.Fprintf(c.Writer, "%+v\n", input)
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
