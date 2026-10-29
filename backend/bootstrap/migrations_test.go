package bootstrap_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"lesson06/backend/bootstrap"
	"lesson06/backend/config"

	"github.com/go-sql-driver/mysql"
)

// TestRunMySQL 验证新环境初始化和重复启动，不连接 DB_NAME 指定的业务库。
// 测试使用独立随机库名，结束时仅删除本次创建的测试库。
func TestRunMySQL(t *testing.T) {
	if os.Getenv("RUN_MYSQL_MIGRATION_TEST") != "1" {
		t.Skip("设置 RUN_MYSQL_MIGRATION_TEST=1 后运行真实 MySQL 迁移测试")
	}
	settings := config.MySQLConfig{
		Addr:     os.Getenv("DB_ADDR"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: fmt.Sprintf("lesson06_migration_test_%d", time.Now().UnixNano()),
	}
	if settings.Addr == "" {
		settings.Addr = "127.0.0.1:3306"
	}
	if settings.User == "" {
		settings.User = "root"
	}
	connection := mysql.NewConfig()
	connection.User = settings.User
	connection.Passwd = settings.Password
	connection.Net = "tcp"
	connection.Addr = settings.Addr
	connection.Timeout = 5 * time.Second
	connection.ReadTimeout = 5 * time.Second
	connection.WriteTimeout = 5 * time.Second
	admin, err := sql.Open("mysql", connection.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE IF EXISTS `" + settings.Database + "`"); err != nil {
			t.Errorf("清理测试数据库 %s: %v", settings.Database, err)
		}
		admin.Close()
	})

	productSeedSQL, err := os.ReadFile("../migration/001_products_seed.sql")
	if err != nil {
		t.Fatal(err)
	}
	taskSeedSQL, err := os.ReadFile("../migration/002_tasks_seed.sql")
	if err != nil {
		t.Fatal(err)
	}
	// 从真实应用入口验证连接、建库、两组迁移及业务依赖装配。
	cfg := &config.Config{MySQL: settings, AdminPassword: "migration-test-password"}
	app, err := bootstrap.InitializeApp(cfg, string(productSeedSQL), string(taskSeedSQL))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { app.SQLDB.Close() })
	db := app.SQLDB
	count := func(table string, want int) {
		t.Helper()
		var got int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s 数量 = %d，期望 %d", table, got, want)
		}
	}
	count("products", 36)
	count("tasks", 9)
	count("lesson_migrations", 2)
	var initialVersion int
	if err := db.QueryRow("SELECT version FROM task_board_meta WHERE id=1").Scan(&initialVersion); err != nil || initialVersion != 1 {
		t.Fatalf("首次看板版本 = %d，错误 = %v", initialVersion, err)
	}

	// 模拟用户修改、删除种子数据以及产生新的看板版本。
	for _, statement := range []string{
		"UPDATE products SET name='保留商品修改', stock=0, status=0, description='' WHERE sku='SP-0001'",
		"DELETE FROM products WHERE sku='SP-0002'",
		"UPDATE tasks SET title='保留任务修改', due_date='' WHERE id=1",
		"DELETE FROM tasks WHERE id=2",
		"UPDATE task_board_meta SET version=42 WHERE id=1",
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	restarted, err := bootstrap.InitializeApp(cfg, string(productSeedSQL), string(taskSeedSQL))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { restarted.SQLDB.Close() })
	count("products", 35)
	count("tasks", 8)
	count("lesson_migrations", 2)

	var name, description string
	var stock, status int
	if err := db.QueryRow("SELECT name, stock, status, description FROM products WHERE sku='SP-0001'").Scan(&name, &stock, &status, &description); err != nil {
		t.Fatal(err)
	}
	if name != "保留商品修改" || stock != 0 || status != 0 || description != "" {
		t.Fatal("重复启动覆盖了商品修改")
	}
	var title, dueDate string
	if err := db.QueryRow("SELECT title, due_date FROM tasks WHERE id=1").Scan(&title, &dueDate); err != nil {
		t.Fatal(err)
	}
	if title != "保留任务修改" || dueDate != "" {
		t.Fatal("重复启动覆盖了任务修改或日期")
	}
	var version int
	if err := db.QueryRow("SELECT version FROM task_board_meta WHERE id=1").Scan(&version); err != nil || version != 42 {
		t.Fatalf("重复启动后的看板版本 = %d，错误 = %v", version, err)
	}
}
