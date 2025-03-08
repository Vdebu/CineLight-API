package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// 用于折嵌套Json 使用key匹配interface{}数据
type envelop map[string]interface{}

// 将用于提取id的代码逻辑抽象成一个模块
func (app *application) readIDParam(c *gin.Context) (int64, error) {
	params := c.Param("id")
	// 提取数字并按照制定格式转化
	id, err := strconv.ParseInt(params, 10, 64)
	if err != nil || id < 1 {
		return 0, err
	}
	return id, nil
}

// 发送json数据到响应体 将data的类型更改为自定义类型用于折叠Json
func (app *application) writeJson(c *gin.Context, status int, data envelop, header http.Header) {
	// 使json的输出更加美观
	c.IndentedJSON(status, data)
	// 向响应体中添加传入的表头
	for key, val := range header {
		// 将 []string 转换为 string，多个值以逗号分隔
		c.Header(key, strings.Join(val, ","))
	}
	//app.logger.Println("we are here...")
}

func (app *application) readJSON(c *gin.Context, dst interface{}) error {
	err := c.BindJSON(dst)
	if err != nil {
		// 尝试判断错误的类型
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmashalError *json.InvalidUnmarshalError
		// 使用errors.As判断返回的错误是否包含特定类型
		switch {
		case errors.As(err, &syntaxError):
			// 检查是否有语法错误并报告错误位置
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			// 检查JSON的类型是否与写入位置的类型不符 并具体判断是否是某个字段不符
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			// 有些情况下也会返回io.ErrunexpectedEOF
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &invalidUnmashalError):
			// 传入的是空指针 直接panic
			panic(err)
		case errors.As(err, &io.EOF):
			// 响应体是空值
			return errors.New("body must not be empty")
		default:
			// 不是上述的任一种类型
			return err
		}
	}
	return nil
}
