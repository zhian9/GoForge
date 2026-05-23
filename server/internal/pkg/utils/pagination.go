package utils

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// Validate 验证并修正分页参数
func (p *Pagination) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算偏移量
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit 获取限制数量
func (p *Pagination) Limit() int {
	return p.PageSize
}

// PageResponse 分页响应
type PageResponse struct {
	Page       int   `json:"page"`        //当前页码
	PageSize   int   `json:"page_size"`   //每页数量
	Total      int64 `json:"total"`       //总记录数
	TotalPages int   `json:"total_pages"` // 总页数
}

// NewPageResponse 创建分页响应
func NewPageResponse(page, pageSize int, total int64) *PageResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PageResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
