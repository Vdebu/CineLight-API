package main

import "github.com/gin-gonic/gin"

func (app *application) routers() *gin.Engine {
	// 创建复用路由 gin.Default默认包含了Logger And Recovery
	router := gin.New()
	router.HandleMethodNotAllowed = true
	// 使用路由分组 会自动以/v1作为前缀
	v1 := router.Group("/v1")
	// 使用中间件 执行顺序-> 确保日志记录在身份验证之前正确捕获身份认证错误的响应体信息
	v1.Use(app.recoverPanic(), app.enableCORS(), app.rateLimiter(), gin.Logger(), app.authenticate())
	{

		v1.GET("/healthcheck", app.healthcheckHandler)
		// 用户相关
		v1.POST("/users", app.registerUserHandler)
		v1.PUT("/users/activated", app.activateUserHandler)
		// gin默认没有为处理器注册OPTIONS方法 需要进行显示处理
		// 添加空的OPTIONS处理方法仅用于处理预检请求
		v1.OPTIONS("/tokens/authentication", func(c *gin.Context) {
			// 空处理器，仅由中间件处理 CORS 头
		})
		// 激活账号
		v1.POST("/tokens/authentication", app.createAuthenticationTokenHandler)
		// 权限敏感的路由组
		private := v1.Group("")
		// 先判断是否认证(登录)再判断是否激活
		private.Use(app.requireAuthenticatedUser(), app.requireActivatedUser())
		{
			// 創建新的路由組 添加檢測用戶權限的中間件(讀寫)
			movies := private.Group("")
			{
				// 行中的中間件執行順序同樣遵從 --->
				movies.POST("/movies", app.requirePermission("movie:write"), app.createMovieHandler)
				movies.GET("/movies/:id", app.requirePermission("movie:read"), app.showMovieHandler)
				// 使用PATCH方法更新信息(一般全部更新用PUT)
				movies.PATCH("/movies/:id", app.requirePermission("movie:write"), app.updateMovieHandler)
				movies.DELETE("/movies/:id", app.requirePermission("movie:write"), app.deleteMovieHandler)
				movies.GET("/movies", app.requirePermission("movie:read"), app.listMoviesHandler)
			}
		}
	}
	return router
}
