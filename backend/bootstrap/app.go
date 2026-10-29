package bootstrap

import (
	"database/sql"

	"lesson06/backend/config"
	authhandler "lesson06/backend/internal/auth/handler"
	authservice "lesson06/backend/internal/auth/service"
	producthandler "lesson06/backend/internal/product/admin/handler"
	productservice "lesson06/backend/internal/product/admin/service"
	"lesson06/backend/internal/product/repository"
	taskhandler "lesson06/backend/internal/task/admin/handler"
	taskservice "lesson06/backend/internal/task/admin/service"
	taskrepository "lesson06/backend/internal/task/repository"

	"gorm.io/gorm"
)

type App struct {
	DB       *gorm.DB
	SQLDB    *sql.DB
	Sessions *authservice.AuthService
	Auth     *authhandler.Handler
	Product  *producthandler.Handler
	Task     *taskhandler.Handler
}

// InitializeApp 与参考项目一致，在启动处显式装配依赖。
func InitializeApp(cfg *config.Config, productSeedSQL, taskSeedSQL string) (*App, error) {
	db, err := LaunchDatabase(&cfg.MySQL)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// 数据库迁移完成后再装配业务模块，确保新环境具备所需表结构。
	if err = initialize(sqlDB, productSeedSQL); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err = initializeTasks(sqlDB, taskSeedSQL); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if err = initializeCategories(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}
	sessions := authservice.NewAuthService(cfg.AdminPassword)
	repo := repository.NewProductRepository(db)
	products := productservice.NewProductService(repo)
	return &App{DB: db, SQLDB: sqlDB, Sessions: sessions, Auth: authhandler.New(sessions), Product: producthandler.New(products), Task: taskhandler.New(taskservice.New(taskrepository.New(db)))}, nil
}
