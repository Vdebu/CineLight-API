package data

import (
	"database/sql"
	"github.com/lib/pq"
	"greenlight.vdebu.net/internal/validator"
	"time"
)

// 使用结构体存储基本信息
type Movie struct {
	ID        int64     `json:"id"`                // 唯一标识
	CreatedAt time.Time `json:"-"`                 // 加入数据库的时间 对数据库中的创建时间字段进行隐藏
	Title     string    `json:"title"`             // 标题
	Year      int32     `json:"year,omitempty"`    // 发行时间
	Runtime   Runtime   `json:"runtime,omitempty"` // 时长 使用自定义类型存储播放时长(实现了json.Marshal接口)生成自定义的格式化信息
	Genres    []string  `json:"genres,omitempty"`  // 标签 对发行时间时长标签进行空值隐藏("",0,nil或空slice,map)
	Version   int32     `json:"version"`           // 版本信息从1开始 当电影信息更新版本信息会自动递增
}

// 用于检测电影结构体的各个字段是否有效
func ValidateMovie(v *validator.Validator, movie *Movie) {
	// 对输入的各个字段进行检查
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "year must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	// 查看是否有标签是重复的
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}

// 创建模型存储数据库连接池
type MovieModel struct {
	db *sql.DB
}

// 向数据库插入新数据
func (m *MovieModel) Insert(movie *Movie) error {
	// 插入的SQL语句
	// $N是Postgresql 特殊占位符
	stmt := `INSERT INTO movies (title,year,runtime,genres)
				VALUES ($1,$2,$3,$4)
				RETURNING id,created_at,version` // 提取数据库自动生成的信息写入传入的结构体
	// 插入三个以上的数据(三个以上的占位符) 使用[]interface{}进行存储并作为参数传入 注意进行类型转换
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	// 将RETURNING返回的数据插入传进来的数据(默认为空值)
	return m.db.QueryRow(stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// 使用id从数据库中查找数据
func (m *MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// 根据新数据更新数据库
func (m *MovieModel) Update(movie *Movie) error {
	return nil
}

// 根据id从数据库中删除数据
func (m *MovieModel) Delete(id int64) error {
	return nil
}
