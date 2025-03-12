package main

import "github.com/gin-gonic/gin"

func (app *application) routers() *gin.Engine {
	// 创建复用路由
	router := gin.Default()
	router.HandleMethodNotAllowed = true
	// 使用路由分组 会自动以/v1作为前缀
	v1 := router.Group("/v1")
	{
		// 使用gin的处理器定义逻辑创建路由
		v1.GET("/healthcheck", app.healthcheckHandler)
		v1.POST("/movies", app.createMovieHandler)
		v1.GET("/movies/:id", app.showMovieHandler)
		// 使用PATCH方法更新信息(一般全部更新用PUT)
		v1.PATCH("/movies/:id", app.updateMovieHandler)
		v1.DELETE("/movies/:id", app.deleteMovieHandler)
		v1.GET("/movies", app.listMoviesHandler)
	}
	return router
}
