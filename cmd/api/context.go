package main

import (
	"github.com/gin-gonic/gin"
	"greenlight.vdebu.net/internal/data"
)

// 使用自定义类型存储进req.context

// 将user(string) -> user(contextKey)避免冲突的同时方便后续进行类型断言
const userContextKey = "user"

// 返回包含新的context的*http.request(将提供的user结构体嵌入请求体的context)
func (app *application) contextSetUser(c *gin.Context, user *data.User) {
	// 提取当前请求的context创建的新的context
	// 复制一个用于覆盖原先的context而不是直接向里面插入新值
	//ctx := context.WithValue(r.Context(), userContextKey, user)

	// 在gin中以上操作方法无法正确将数据插入context 需要使用gin的内置方法
	// 直接插入自定义的userContextKey
	c.Set(userContextKey, user)
	// 即插入成功后续也无需return
	// 返回覆盖了新context的请求体
	//return r.WithContext(ctx)
}

// 从r.ctx中提取user结构体 -> 只会在当前请求的ctx中应该会有user的时候调用如果没提取到直接panic
func (app *application) contextGetUser(c *gin.Context) *data.User {
	// 这里是非接口类型断言也无效
	val, ok := c.Get(userContextKey)
	// 先判断是否存在对应的键
	if !ok {
		panic("missing user value in request context")
	}
	// 尝试进行类型断言
	user, ok := val.(*data.User)
	if !ok {
		// 若不是预测的类型直接panic
		panic("wrong authentication key type")
	}
	return user
}
