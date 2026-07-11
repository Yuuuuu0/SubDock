package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"subdock/internal/config"
	"subdock/internal/model"
	"subdock/internal/router"
	"subdock/internal/scheduler"
)

// main 启动 SubDock，并在启动或运行失败时输出明确错误。
func main() {
	if err := run(); err != nil {
		log.Fatalf("SubDock 运行失败: %v", err)
	}
}

// run 组装应用依赖、启动 HTTP 与调度服务，并处理优雅关闭。
func run() error {
	log.Println("SubDock starting...")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	database, err := model.InitDB()
	if err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %w", err)
	}
	defer sqlDB.Close()

	sched := scheduler.New(database)
	if err := sched.Start(); err != nil {
		return err
	}
	defer sched.Stop()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router.Setup(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("服务器启动在 http://localhost%s", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("关闭 HTTP 服务失败: %w", err)
		}
		return nil
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("启动 HTTP 服务失败: %w", err)
	}
}
