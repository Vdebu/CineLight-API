package data

import (
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
