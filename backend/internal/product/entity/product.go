package entity

import "time"

// Product 对应现有 products 表；迁移由 bootstrap 显式管理。
type Product struct {
	ID          int64 `gorm:"primaryKey;autoIncrement"`
	Name        string
	SKU         string `gorm:"column:sku"`
	Category    string
	PriceCents  int64
	Stock       int
	Status      int
	Description string
	UpdatedAt   time.Time `gorm:"autoUpdateTime:false;default:CURRENT_TIMESTAMP"`
}

func (Product) TableName() string { return "products" }
