package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lesson06/backend/bootstrap"
	"lesson06/backend/config"
	"lesson06/backend/router"

	"github.com/gin-gonic/gin"
)

// SQL 编译进程序，运行时无需从工作目录读取文件。
//
//go:embed migration/001_products_seed.sql
var productSeedSQL string

//go:embed migration/002_tasks_seed.sql
var taskSeedSQL string

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	app, err := bootstrap.InitializeApp(cfg, productSeedSQL, taskSeedSQL)
	if err != nil {
		log.Fatal("application initialization failed: ", err)
	}
	defer app.SQLDB.Close()
	gin.SetMode(gin.ReleaseMode)
	server := &http.Server{Addr: cfg.Server.Addr, Handler: router.SetupRouter(app), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	fmt.Printf("Lesson 06 Gin API: http://%s\n", server.Addr)
	if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
