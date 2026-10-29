package dto

// Status 使用指针区分未填写与合法的下架状态 0。
type SaveRequest struct {
	Name        string `json:"name"`
	SKU         string `json:"sku"`
	CategoryID  int64  `json:"categoryId"`
	PriceCents  int64  `json:"priceCents"`
	Stock       int    `json:"stock"`
	Status      *int   `json:"status"`
	Description string `json:"description"`
}
type ListQuery struct {
	Keyword    string `form:"keyword"`
	CategoryID int64  `form:"categoryId"`
	Status     string `form:"status"`
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"pageSize,default=10"`
}
type BatchStatusRequest struct {
	IDs    []int64 `json:"ids"`
	Status *int    `json:"status"`
}
type Product struct {
	CategoryName string `json:"categoryName"`
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	SKU          string `json:"sku"`
	CategoryID   int64  `json:"categoryId"`
	PriceCents   int64  `json:"priceCents"`
	Stock        int    `json:"stock"`
	Status       int    `json:"status"`
	Description  string `json:"description"`
	UpdatedAt    string `json:"updatedAt"`
}
type PageResult struct {
	Items    []Product `json:"items"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}
