package data

import "greenlight.vdebu.net/internal/validator"

// 存储其他请求端点也可能用上的字段信息
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

// 检查filters字段的有效性
func ValidateFilters(v *validator.Validator, f Filters) {
	// 检查 pages
	v.Check(f.Page > 0, "page", "must greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "pages_size", "must greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	// 检查sort的key是否都在允许的范围内
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}
