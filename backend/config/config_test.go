package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadValidation(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ADMIN_PASSWORD", "123456")
	t.Setenv("DB_NAME", "lesson06_shop")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DB_NAME", "bad-name;")
	if _, err := Load(); err == nil {
		t.Fatal("invalid database name accepted")
	}
	t.Setenv("DB_NAME", "lesson06_shop")
	t.Setenv("ADMIN_PASSWORD", "short")
	if _, err := Load(); err == nil {
		t.Fatal("short password accepted")
	}
}

func TestLoadDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, key := range []string{"HTTP_ADDR", "DB_ADDR", "DB_USER", "DB_PASSWORD", "DB_NAME", "ADMIN_PASSWORD"} {
		t.Setenv(key, "") // 测试结束后恢复原环境。
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	content := "# 本地配置\nHTTP_ADDR=127.0.0.1:18080\nDB_ADDR=127.0.0.1:3307\nDB_USER=student\nDB_NAME=lesson_test\nDB_PASSWORD='pass # with = symbols'\nADMIN_PASSWORD=class-password\n"
	if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Addr != "127.0.0.1:18080" || cfg.MySQL.Addr != "127.0.0.1:3307" || cfg.MySQL.User != "student" || cfg.MySQL.Database != "lesson_test" || cfg.MySQL.Password != "pass # with = symbols" || cfg.AdminPassword != "class-password" {
		t.Fatal("未正确读取 .env 配置")
	}
	if _, exists := os.LookupEnv("DB_PASSWORD"); exists {
		t.Fatal("读取配置不应修改进程环境")
	}
	t.Setenv("DB_NAME", "environment_db")
	t.Setenv("DB_PASSWORD", "")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MySQL.Database != "environment_db" || cfg.MySQL.Password != "" {
		t.Fatal("系统环境变量（包括显式空值）应优先于 .env")
	}
}

func TestLoadInvalidDotEnv(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ADMIN_PASSWORD", "test-password")
	if err := os.WriteFile(".env", []byte("DB_PASSWORD='secret-unclosed"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), ".env") || strings.Contains(err.Error(), "secret-unclosed") {
		t.Fatal("格式错误应阻止启动，且不泄露配置原文")
	}
}
