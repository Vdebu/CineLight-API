package main

import "github.com/gin-gonic/gin"

func (app *application) routers() *gin.Engine {
	// 创建复用路由
	router := gin.Default()
	// 使用gin的处理器定义逻辑创建路由
	router.GET("/v1/healthcheck", app.healthcheckHandler)
	router.POST("/v1/movies", app.createMovieHandler)
	router.GET("/v1/movies/:id", app.showMovieHandler)
	return router
}
