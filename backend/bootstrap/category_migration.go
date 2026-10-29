package bootstrap

import "database/sql"

// initializeCategories 将旧分类名称迁移成关联 ID。DDL 可重复执行，迁移中断后可重试。
func initializeCategories(db *sql.DB) error {
	var done int
	if err := db.QueryRow("SELECT COUNT(*) FROM lesson_migrations WHERE version=3").Scan(&done); err != nil {
		return err
	}
	if done > 0 {
		return nil
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS categories (
 id BIGINT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(32) NOT NULL UNIQUE
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return err
	}
	if _, err := db.Exec("INSERT IGNORE INTO categories(id,name) VALUES (1,'数码配件'),(2,'办公文具'),(3,'生活好物')"); err != nil {
		return err
	}
	var exists int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name='products' AND column_name='category_id'").Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		if _, err := db.Exec("ALTER TABLE products ADD COLUMN category_id BIGINT NULL, ADD INDEX idx_products_category_id(category_id)"); err != nil {
			return err
		}
	}
	var indexed int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name='products' AND index_name='idx_products_category_id'").Scan(&indexed); err != nil {
		return err
	}
	if indexed == 0 {
		if _, err := db.Exec("ALTER TABLE products ADD INDEX idx_products_category_id(category_id)"); err != nil {
			return err
		}
	}
	// 保留旧库中可能存在的自定义分类，不丢失原有商品归属。
	if _, err := db.Exec("INSERT IGNORE INTO categories(name) SELECT DISTINCT category FROM products WHERE category <> ''"); err != nil {
		return err
	}
	if _, err := db.Exec("UPDATE products p JOIN categories c ON c.name=p.category SET p.category_id=c.id, p.updated_at=p.updated_at WHERE p.category_id IS NULL"); err != nil {
		return err
	}
	// 旧 category 列仅供历史种子文件兼容；业务不再读写它。
	if _, err := db.Exec("ALTER TABLE products MODIFY category_id BIGINT NOT NULL, MODIFY category VARCHAR(32) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	_, err := db.Exec("INSERT IGNORE INTO lesson_migrations(version) VALUES(3)")
	return err
}
