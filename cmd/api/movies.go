package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"greenlight.vdebu.net/internal/data"
	"greenlight.vdebu.net/internal/validator"
	"net/http"
)

func (app *application) createMovieHandler(c *gin.Context) {
	// 创建结构体存储待解析的JSON数据 这个结构体是完整Movie结构体的子集
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"` // 使用自定义数据存储时间
		Genres  []string     `json:"genres"`
	}
	err := app.readJSON(c, &input)
	if err != nil {
		// 向响应体中输出错误信息
		app.badRequestResponse(c, err)
		return
	}
	// 数据提取成功后初始化校验器 以后可能会有各种各样的校验器
	v := validator.New()
	// 将输入的信息载入Movie结构体用于后续检测(这里创建的是movie指针 用于后续填写自动生成的信息)
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	// 尝试将用户输入的信息插入到数据库中
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	// 插入成功后向响应头写入数据展示新信息的存储位置
	headers := make(http.Header)
	// 设置Location属性
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))
	// 返回创建好的JSON信息
	app.writeJson(c, http.StatusOK, envelop{"movie": movie}, headers)
}
func (app *application) showMovieHandler(c *gin.Context) {
	// 使用抽象出来的数据读取模块提取id
	id, err := app.readIDParam(c)
	if err != nil {
		app.notFoundResponse(c)
		return
	}
	// 从数据库查找数据
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		// 判断错误类型是否是无记录
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(c)
		} else {
			app.serverErrorResponse(c, err)
		}
		return
	}
	// 将电影的信息以json的形式输出 使用自定义类型进行封装以呈现出嵌套展示的效果
	app.writeJson(c, http.StatusOK, envelop{"movie": movie}, nil)
}
