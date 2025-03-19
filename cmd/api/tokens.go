package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"greenlight.vdebu.net/internal/data"
	validator2 "greenlight.vdebu.net/internal/validator"
	"net/http"
	"time"
)

// 读取用户输入的邮箱与密码生成相关的认证秘钥
func (app *application) createAuthenticationTokenHandler(c *gin.Context) {
	// 读取用户输入的账号密码
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(c, &input)
	if err != nil {
		// 读取失败输入的JSON字段有误
		app.badRequestResponse(c, err)
		return
	}
	// 预先进行简单的有效性检查
	v := validator2.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	// 尝试使用给定的信息从数据库提取用户
	user, err := app.models.User.GetByEmail(input.Email)
	if err != nil {
		// 判断错误类型
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// 输入的信息有误
			app.invalidCredentialResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	// 正确提取到了对应邮箱的用户信息开始判断密码的有效性
	match, err := user.Password.Matches(input.Password)
	// 先判断是否发生了错误(错误可能是由生成哈希造成的)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	// 没有发生错误再检查是否匹配成功
	if !match {
		// 输入的信息无效
		app.invalidCredentialResponse(c)
		return
	}
	// 邮箱与密码都是匹配的则进行认证秘钥的生成
	// 绑定当前用户的ID时限为一天作用域为认证
	token, err := app.models.Token.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	// 将生成的token信息返回给用户
	app.writeJson(c, http.StatusCreated, envelop{"token": token}, nil)
}
