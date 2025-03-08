package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 定义用于在解析失败时返回的错误
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Runtime 创建自定义类型存储时长 依赖于int32
type Runtime int32

// MarshalJSON 在自定义类型上实现json.MarshalJSON接口
func (r Runtime) MarshalJSON() ([]byte, error) {
	// 生成自定义的字符串描述播放时长
	// 使用r对时长进行格式化
	jsonValue := fmt.Sprintf("%d mins", r)
	// 生成的自定义字符串使用""进行包裹使其成为有效的JSON string
	quotedJSONValue := strconv.Quote(jsonValue)
	// 将自定格式化的结果返回
	return []byte(quotedJSONValue), nil
}

// 实现UnMarshalJSON接口 由于UnMarshalJSON会对值进行修改 所以必须是指针接收器
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// 传入的应是一个字符串 "<runtime> mins" 先把双引号去掉
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	// 以空格为基准将字符串分割
	parts := strings.Split(unquotedJSONValue, " ")
	// 应该是被分割成了两份 并且最后一份表示的是分钟
	if len(parts) != 2 || parts[1] != "mins" {
		// 与预期情况不符解析错误
		return ErrInvalidRuntimeFormat
	}
	// 将string数字转换为int32
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	// 转换成功修改传入的原始Runtime
	*r = Runtime(i)
	return nil
}
