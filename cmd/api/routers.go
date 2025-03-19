package main

import "github.com/gin-gonic/gin"

func (app *application) routers() *gin.Engine {
	// 创建复用路由 gin.Default默认包含了Logger And Recovery
	router := gin.New()
	router.HandleMethodNotAllowed = true
	// 使用路由分组 会自动以/v1作为前缀
	v1 := router.Group("/v1")
	// 使用中间件 执行顺序-> 确保日志记录在身份验证之前正确捕获身份认证错误的响应体信息
	v1.Use(app.recoverPanic(), app.rateLimiter(), gin.Logger(), app.authenticate())
	{
		// 使用gin的处理器定义逻辑创建路由
		v1.GET("/healthcheck", app.healthcheckHandler)
		v1.POST("/movies", app.createMovieHandler)
		v1.GET("/movies/:id", app.showMovieHandler)
		// 使用PATCH方法更新信息(一般全部更新用PUT)
		v1.PATCH("/movies/:id", app.updateMovieHandler)
		v1.DELETE("/movies/:id", app.deleteMovieHandler)
		v1.GET("/movies", app.listMoviesHandler)
		// 用户相关
		v1.POST("/users", app.registerUserHandler)
		v1.PUT("/users/activated", app.activateUserHandler)
		// 激活账号
		v1.POST("/tokens/authentication", app.createAuthenticationTokenHandler)
	}
	return router
}
