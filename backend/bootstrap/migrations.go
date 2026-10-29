package bootstrap

import "database/sql"

// initialize 初始化商品表，迁移版本 1 仅首次插入 36 条演示商品。
func initialize(db *sql.DB, seedSQL string) error {
	// 金额以整数分保存；数据库约束与接口的字段范围保持一致。
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS products (
 id BIGINT PRIMARY KEY AUTO_INCREMENT,
 name VARCHAR(80) NOT NULL,
 sku VARCHAR(32) NOT NULL UNIQUE,
 category VARCHAR(32) NOT NULL,
 price_cents BIGINT NOT NULL,
 stock INT NOT NULL,
 status TINYINT NOT NULL,
 description VARCHAR(500) NOT NULL DEFAULT '',
 updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
 CHECK (price_cents BETWEEN 1 AND 99999999), CHECK(stock BETWEEN 0 AND 999999), CHECK(status IN (0,1))
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS lesson_migrations (version INT PRIMARY KEY)`)
	if err != nil {
		return err
	}
	// 版本登记与种子写入使用同一事务，失败后可在下次启动时重试。
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec("INSERT IGNORE INTO lesson_migrations(version) VALUES(1)")
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 1 {
		if _, err := tx.Exec(seedSQL); err != nil {
			return err
		}
	}
	return tx.Commit()
}
