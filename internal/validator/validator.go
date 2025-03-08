package validator

import "regexp"

// EmailRX 使用正则表达式匹配输入的邮箱
var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Validator struct {
	Errors map[string]string
}

// 初始化验证器
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// 检查是否有错误发生
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// 向验证器的错误字典中添加字段
func (v *Validator) AddError(key, message string) {
	// 如果不存在则加入 只显示第一个发生的错误信息
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// 检查有效性 没通过就调用加入字段方法
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// 判断输入的字符串是否在给定列表中
func In(value string, list ...string) bool {
	for i := range list {
		if value == list[i] {
			return true
		}
	}
	return false
}

// 判断输入的字符串是否符合正则表达式
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// 判断字符串是否唯一
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	// 判断记录数是否一致
	return len(uniqueValues) == len(values)
}
