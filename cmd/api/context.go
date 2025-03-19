package main

import (
	"context"
	"greenlight.vdebu.net/internal/data"
	"net/http"
)

// 使用自定义类型存储进req.context
type contextKey string

// 将user(string) -> user(contextKey)避免冲突的同时方便后续进行类型断言
const userContextKey = contextKey("user")

// 返回包含新的context的*http.request(将提供的user结构体嵌入请求体的context)
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	// 提取当前请求的context创建的新的context
	// 复制一个用于覆盖原先的context而不是直接向里面插入新值
	ctx := context.WithValue(r.Context(), userContextKey, user)
	// 返回覆盖了新context的请求体
	return r.WithContext(ctx)
}

// 从r.ctx中提取user结构体 -> 只会在当前请求的ctx中应该会有user的时候调用如果没提取到直接panic
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
