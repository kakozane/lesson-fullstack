package entity

import "time"

// Product 对应现有 products 表；迁移由 bootstrap 显式管理。
type Product struct {
	ID          int64 `gorm:"primaryKey;autoIncrement"`
	Name        string
	SKU         string `gorm:"column:sku"`
	CategoryID  int64
	Category    Category `gorm:"foreignKey:CategoryID"`
	PriceCents  int64
	Stock       int
	Status      int
	Description string
	UpdatedAt   time.Time `gorm:"autoUpdateTime:false;default:CURRENT_TIMESTAMP"`
}

func (Product) TableName() string { return "products" }

// Category 是商品分类字典，商品通过 ID 关联。
type Category struct {
	ID   int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `json:"name"`
}

func (Category) TableName() string { return "categories" }
