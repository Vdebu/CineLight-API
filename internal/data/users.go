package data

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"greenlight.vdebu.net/internal/validator"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// 与数据库中表结构一致的结构体 对密码与当前的版本号信息进行了隐藏
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// 存储密码信息哈希与未哈希的版本
type password struct {
	plaintext *string
	hash      []byte
}

// 用户的数据库连接池模型
type UserModel struct {
	db *sql.DB
}

// 将输入的密码进行哈希并赋值给当前传入的password结构体
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	// 更新当前结构体的数据
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// 检查输入的密码与存储的哈希值是否匹配
func (p *password) Matches(plaintextPassword string) (bool, error) {
	// 比较存储在结构体中的哈希是否与输入密码的哈希匹配
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		// 判断是否是哈希值不匹配造成的错误
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// 不返回真实的错误 -> 预料之中
			return false, nil
		default:
			return false, err
		}
	}
	// 匹配成功
	return true, err
}

// 验证输入的邮箱是否合法
func ValidateEmail(v *validator.Validator, email string) {
	// 检查邮箱是否是空的
	v.Check(email != "", "email", "must be provided")
	// 使用正则表达式检查有效性
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// 检查密码的有效性
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	// 检查是否为空值
	v.Check(password != "", "password", "must be provided")
	// 限制密码的长度防止hash值泄露
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
func ValidateUser(v *validator.Validator, user *User) {
	// 名称不为空
	v.Check(user.Name != "", "name", "must be provided")
	// 名称不能太长
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	// 邮箱的有效性
	ValidateEmail(v, user.Email)
	// 若密码不为nil ->用户进行了输入 对有效性进行判断
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// 如果用户密码的哈希值是空的说明代码出现了逻辑错误 直接panic
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

}
func (m *UserModel) Insert(user *User) error {
	stmt := `
			INSERT INTO users(name,email,password_hash,activated)
			VALUES ($1,$2,$3,$4)
			RETURNING id,created_at,version`
	// 载入需要插入的参数
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}
	// 设置操作的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	// 回收资源
	defer cancel()
	// 将自动生成的信息返回填入User使其更完整
	err := m.db.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		// 检查错误是不是由于重复插入email造成的
		case err.Error() == `pq:duplicate key value violate unique constraint "user_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	// 插入成功
	return nil
}

// 通过提供的email查询用户的具体信息
func (m *UserModel) GetByEmail(email string) (*User, error) {
	stmt := `
			SELECT id,created_at,name,email,password_hash,activated,version
			FROM users
			WHERE email = $1`
	// 创建结构体存储查询到的信息
	var user User
	// 设置五秒操作超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := m.db.QueryRowContext(ctx, stmt, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		// 判断是否查无此人
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// 返回查询到的结果
	return &user, nil
}

// 更新用户的信息(乐观锁)
func (m *UserModel) Update(user *User) error {
	stmt := `
			UPDATE users
			SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
			WHERE id = $5  AND  version = $6
			RETURNING version`
	// 载入用于更新的信息
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}
	// 五秒操作超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 同时更新当前user的version
	err := m.db.QueryRowContext(ctx, stmt, args...).Scan(&user.Version)
	if err != nil {
		switch {
		// 检查错误类型
		case err.Error() == `pq:duplicate key value violate unique constraint "user_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}
