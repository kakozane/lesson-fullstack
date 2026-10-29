package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server        ServerConfig
	MySQL         MySQLConfig
	AdminPassword string
}
type ServerConfig struct{ Addr string }
type MySQLConfig struct{ Addr, User, Password, Database string }

func Load() (*Config, error) {
	fileEnv, err := godotenv.Read(".env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		// 解析错误可能包含原文，不把密码等配置内容写入启动日志。
		return nil, fmt.Errorf("读取 .env 失败，请检查文件权限和格式")
	}
	// 系统环境变量优先；只读取文件，不修改进程环境。
	env := func(key, fallback string) string {
		value, exists := os.LookupEnv(key)
		if !exists {
			value = fileEnv[key]
		}
		if value == "" {
			return fallback
		}
		return value
	}
	cfg := &Config{
		Server:        ServerConfig{Addr: env("HTTP_ADDR", "127.0.0.1:8080")},
		MySQL:         MySQLConfig{Addr: env("DB_ADDR", "127.0.0.1:3306"), User: env("DB_USER", "root"), Password: env("DB_PASSWORD", ""), Database: env("DB_NAME", "lesson06_shop")},
		AdminPassword: env("ADMIN_PASSWORD", ""),
	}
	if len(cfg.AdminPassword) < 6 {
		return nil, fmt.Errorf("ADMIN_PASSWORD must contain at least 6 characters")
	}
	for _, c := range cfg.MySQL.Database {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return nil, fmt.Errorf("DB_NAME only accepts lowercase letters, digits and underscores")
		}
	}
	return cfg, nil
}
