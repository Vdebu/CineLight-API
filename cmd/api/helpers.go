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
	// 通过重新定义gin的请求体限制JSON请求体的大小 防止服务器资源耗尽
	// 这里限制为1mb
	maxBytes := 1 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxBytes))
	// 使用自定义解析器解析JSON数据 Gin本身并没有能防止未知字段的方法
	dec := json.NewDecoder(c.Request.Body)
	// 拒绝未知字段 如果JSON中包含未知字段会直接报错
	dec.DisallowUnknownFields()
	// err := c.BindJSON(dst)
	// 使用自定义解析器解析数据
	err := dec.Decode(dst)
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
			// 有些情况下也会返回io.errUnexpectedEOF
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &invalidUnmashalError):
			// 传入的是空指针 直接panic
			panic(err)
		case errors.Is(err, io.EOF):
			// 响应体是空值
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			// 检查是否包含未知字段的错误前缀
			// 如果包含则将字段名<name>提取并包装为错误返回
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		default:
			// 不是上述的任一种类型
			return err
		}
	}
	// 创建空的struct并再次调用解析器检查是否有额外的数据被输入
	err = dec.Decode(&struct {
	}{})
	if err != io.EOF {
		// 有额外的数据存在返回错误
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}
