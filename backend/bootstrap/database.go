package bootstrap

import (
	"database/sql"
	"time"

	"lesson06/backend/config"

	"github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// LaunchDatabase 创建课程数据库，并初始化 GORM 与底层连接池。
// 数据库名由 config.Load 校验；已有数据库不会被清空或重建。
func LaunchDatabase(settings *config.MySQLConfig) (*gorm.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = settings.User
	cfg.Passwd = settings.Password
	cfg.Net = "tcp"
	cfg.Addr = settings.Addr
	cfg.ParseTime = true
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	cfg.Loc = location
	cfg.ClientFoundRows = true
	cfg.Timeout = 5 * time.Second
	cfg.ReadTimeout = 5 * time.Second
	cfg.WriteTimeout = 5 * time.Second
	cfg.Params = map[string]string{"charset": "utf8mb4", "time_zone": "'+08:00'"}
	// 首先使用不指定数据库的连接建库，便于新环境首次启动。
	name := settings.Database
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if _, err = db.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		return nil, err
	}
	// 后续业务连接仅使用配置指定的课程库。
	cfg.DBName = name
	result, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	result.SetMaxOpenConns(10)
	result.SetMaxIdleConns(5)
	result.SetConnMaxLifetime(3 * time.Minute)
	if err = result.Ping(); err != nil {
		result.Close()
		return nil, err
	}
	gdb, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: result}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		result.Close()
		return nil, err
	}
	return gdb, nil
}
