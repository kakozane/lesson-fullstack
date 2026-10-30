package repository

import (
	"context"
	"database/sql"

	"lesson06/backend/internal/product/entity"

	"gorm.io/gorm"
)

type ProductRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db: db} }
func (r *ProductRepository) ListPaged(ctx context.Context, keyword string, categoryID int64, status string, page, size int) ([]entity.Product, int64, int, error) {
	items := []entity.Product{}
	var total int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&entity.Product{})
		if keyword != "" {
			q = q.Where("LOCATE(?,name)>0 OR LOCATE(?,sku)>0", keyword, keyword)
		}
		if categoryID != 0 {
			q = q.Where("category_id = ?", categoryID)
		}
		if status != "" {
			q = q.Where("status = ?", status)
		}
		if err := q.Count(&total).Error; err != nil {
			return err
		}
		page = min(page, max(1, int((total+int64(size)-1)/int64(size))))
		return q.Order("id DESC").Limit(size).Offset((page - 1) * size).Preload("Category").Find(&items).Error
	}, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	return items, total, page, err
}
func (r *ProductRepository) Create(ctx context.Context, p *entity.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}
func (r *ProductRepository) Update(ctx context.Context, id int64, p *entity.Product) error {
	// map 显式更新零值，确保库存 0、下架 0 和空描述不会被 GORM 忽略。
	result := r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).Updates(map[string]any{
		"name": p.Name, "sku": p.SKU, "category_id": p.CategoryID, "price_cents": p.PriceCents, "stock": p.Stock, "status": p.Status, "description": p.Description, "updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entity.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (r *ProductRepository) UpdateStatus(ctx context.Context, id int64, status int) error {
	result := r.db.WithContext(ctx).Model(&entity.Product{}).Where("id = ?", id).Updates(map[string]any{"status": status, "updated_at": gorm.Expr("CURRENT_TIMESTAMP")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepository) Categories(ctx context.Context) ([]entity.Category, error) {
	items := []entity.Category{}
	err := r.db.WithContext(ctx).Order("id").Find(&items).Error
	return items, err
}
func (r *ProductRepository) HasCategory(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Category{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
