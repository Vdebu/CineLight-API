package data

import (
	"fmt"
	"strconv"
)

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
