package data

// 存储其他请求端点也可能用上的字段信息
type Filters struct {
	Page     int
	PageSize int
	Sort     string
}
