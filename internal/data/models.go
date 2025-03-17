package data

import (
	"database/sql"
	"errors"
)

// 自定义错误类型
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// 存储各种数据模型 这样存储进去不会包含sql.db相当于是将其隐藏了 只会包含方法不会包含原始字段？
type Models struct {
	Movies MovieModel
	User   UserModel
	Token  TokenModel
}

// 创建新的模型实例
func NewModels(db *sql.DB) Models {
	return Models{
		// 初始化数据模型的数据库连接池
		Movies: MovieModel{db: db},
		User:   UserModel{db: db},
		Token:  TokenModel{db: db},
	}
}
