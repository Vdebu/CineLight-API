package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"greenlight.vdebu.net/internal/data"
	validator2 "greenlight.vdebu.net/internal/validator"
	"net/http"
)

func (app *application) registerUserHandler(c *gin.Context) {
	// 提取用户输入的数据
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(c, &input)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}
	// 提取输入的信息
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false, // 默认状态处于未激活
	}
	// 使用set方法进行密码的设置
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	// 创建验证器用于验证有效性
	v := validator2.New()
	// 检验输入信息的有效性
	if data.ValidateUser(v, user); !v.Valid() {
		// 输出错误信息
		app.failedValidationResponse(c, v.Errors)
		return
	}
	// 数据准确尝试向数据库插入
	err = app.models.User.Insert(user)
	if err != nil {
		// 判断错误类型
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(c, v.Errors)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	// 展示成功创建的信息
	app.writeJson(c, http.StatusCreated, envelop{"user": user}, nil)
}
