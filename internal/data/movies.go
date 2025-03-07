package data

import "time"

// 使用结构体存储基本信息
type Movie struct {
	ID        int64     `json:"id"`                // 唯一标识
	CreatedAt time.Time `json:"-"`                 // 加入数据库的时间 对数据库中的创建时间字段进行隐藏
	Title     string    `json:"title"`             // 标题
	Year      int32     `json:"year,omitempty"`    // 发行时间
	Runtime   int32     `json:"runtime,omitempty"` // 时长 对发行时间时长标签进行空值隐藏("",0,nil或空slice,map)
	Genres    []string  `json:"genres,omitempty"`  // 标签
	Version   int32     `json:"version"`           // 版本信息从1开始 当电影信息更新版本信息会自动递增
}
