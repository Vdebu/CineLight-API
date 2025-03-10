package data

import (
	"database/sql"
	"errors"
)

// 自定义错误类型
var (
	ErrRecordNotFound = errors.New("record not found")
)

// 存储各种数据模型 这样存储进去不会包含sql.db相当于是将其隐藏了 只会包含方法不会包含原始字段？
type Models struct {
	Movies MovieModel
}

// 创建新的模型实例
func NewModels(db *sql.DB) Models {
	return Models{Movies: MovieModel{db: db}}
}
