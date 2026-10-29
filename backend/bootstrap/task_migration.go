package bootstrap

import "database/sql"

// initializeTasks 初始化任务表和看板版本，迁移版本 2 仅首次插入 9 条任务。
// 必须在商品迁移之后执行，复用已经创建的 lesson_migrations 版本表。
func initializeTasks(db *sql.DB, seedSQL string) error {
	// 看板版本用于任务写入的冲突检测，已有版本值不能被启动流程重置。
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS tasks (
   id BIGINT PRIMARY KEY AUTO_INCREMENT, title VARCHAR(80) NOT NULL,
   description VARCHAR(2000) NOT NULL DEFAULT '', assignee VARCHAR(30) NOT NULL DEFAULT '',
   priority VARCHAR(10) NOT NULL, status VARCHAR(10) NOT NULL,
   due_date VARCHAR(10) NOT NULL DEFAULT '', position INT NOT NULL,
   updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   INDEX idx_task_order(status,position),
   CHECK(status IN ('todo','doing','done')), CHECK(priority IN ('low','medium','high'))
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS task_board_meta (id INT PRIMARY KEY, version BIGINT NOT NULL) ENGINE=InnoDB`,
		`INSERT IGNORE INTO task_board_meta(id,version) VALUES(1,1)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	// 种子日期以首次初始化当天为基准；重启后不刷新日期或补回已删除任务。
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec("INSERT IGNORE INTO lesson_migrations(version) VALUES(2)")
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
