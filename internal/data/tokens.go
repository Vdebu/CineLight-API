package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"greenlight.vdebu.net/internal/validator"
	"time"
)

// 定义tokens可以作用的范围
var (
	ScopeActivation = "activation"
)

// 定义结构体用于存储Token的相关信息
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

// Token数据模型解耦数据库相关的操作
type TokenModel struct {
	db *sql.DB
}

// 生成Token
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// 预先创建Token的基础信息
	token := &Token{
		UserID: userID,              // 用户ID信息
		Expiry: time.Now().Add(ttl), // 根据现在的时间计算过期时间
		Scope:  scope,               // 可作用范围
	}
	// 创建空间为16的[]byte存储随机生成的bytes
	randomBytes := make([]byte, 16)
	// 使用Read方法将切片填充满随机生成的bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	// 将随机填充的bytes转换成字符串
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	// 生成SHA-256的哈希
	hash := sha256.Sum256([]byte(token.Plaintext))
	// 返回的是一个长度为32的数组 将其转换成切片
	token.Hash = hash[:]
	// 返回生成好的token
	return token, nil
}

// 对未哈希的密码进行基础检测
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

// 创建新的Token通过调用GenerateToken实现
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	// 创建成功向数据库插入新的Token
	err = m.Insert(token)
	// 没有其他操作不需要判断错误类型直接返回
	return token, err
}

// 向数据库中插入新的Token
func (m TokenModel) Insert(token *Token) error {
	stmt := `
			INSERT INTO tokens(hash,user_id,expiry,scope)
			VALUES ($1,$2,$3,$4)`
	// 载入要插入的参数
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	// 设置五秒的操作超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 回收资源
	defer cancel()
	// 执行插入操作
	_, err := m.db.ExecContext(ctx, stmt, args)
	// 后续没有操作了 直接返回
	return err
}

// 针对某个用户删除其所有的Token
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	stmt := `
			DELETE FROM users
			WHERE scope = $1 AND user_id = $2`
	// 设置五秒操作超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 回收资源
	defer cancel()
	_, err := m.db.ExecContext(ctx, stmt, scope, userID)
	return err
}
