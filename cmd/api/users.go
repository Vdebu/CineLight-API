package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"greenlight.vdebu.net/internal/data"
	validator2 "greenlight.vdebu.net/internal/validator"
	"net/http"
	"time"
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
	// 為新創建的賬號設置讀權限
	err = app.models.Permissions.AddForUser(user.ID, "movie:read")
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	// 生成用于账号激活的Token
	// 指定当前数据库生成的userID时效为三天范围仅限激活
	token, err := app.models.Token.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	app.background(func() {
		// 使用字典存储要嵌入邮件的数据
		emailData := map[string]interface{}{
			"userID":          user.ID,         // 注册成功的用户ID
			"activationToken": token.Plaintext, // 未哈希用于激活账号的Token
		}
		// 使用goroutine完成邮件的发送操作节省完成请求所需要的时间
		// 注册成功后向用户的邮箱发送欢迎邮件
		err = app.mailer.Send(user.Email, "user_welcome.tmpl.html", emailData)
		// 这里不能使用app.serverErrorResponse 因为在这之前服务器可能已经正确处理请求写入响应体
		// 使用app.logger.PrintError输出错误信息
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	// 展示成功创建的信息
	app.writeJson(c, http.StatusCreated, envelop{"user": user}, nil)
}

func (app *application) activateUserHandler(c *gin.Context) {
	// 存储用户输入的JSON
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	// 尝试进行读取
	err := app.readJSON(c, &input)
	if err != nil {
		// 输入有误 badRequest
		app.badRequestResponse(c, err)
		return
	}
	// 验证输入Token的基本有效性
	v := validator2.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		// 报告具体错误信息
		app.failedValidationResponse(c, v.Errors)
		return
	}
	// 从数据库中检查当前的Token是否有效
	user, err := app.models.User.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		// 检查错误类型
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// 没有找到相关的记录
			v.AddError("token", "invalid or expiry token")
			app.failedValidationResponse(c, v.Errors)
		default:
			// 服务器内部的错误
			app.serverErrorResponse(c, err)
		}
		return
	}
	// 找到了相关的记录将用户的状态设置为已激活
	user.Activated = true
	// 更新数据库中的用户信息
	err = app.models.User.Update(user)
	if err != nil {
		// 检查错误是否是由发生编辑冲突造成的
		switch {
		case errors.Is(err, data.ErrEditConflict):
			// 返回编辑冲突错误
			app.editConflictResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	// 根据当前用户的ID删除使用的Token
	err = app.models.Token.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	// 输出新的用户信息
	app.writeJson(c, http.StatusOK, envelop{"user": user}, nil)
}
