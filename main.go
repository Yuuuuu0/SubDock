package main

import (
	"fmt"
	"log"

	"subdock/internal/config"
	"subdock/internal/model"
	"subdock/internal/router"
	"subdock/internal/scheduler"
)

func main() {
	log.Println("SubDock starting...")

	cfg := config.Load()

	if _, err := model.InitDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	sched := scheduler.New()
	sched.Start()
	defer sched.Stop()

	r := router.Setup()

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("服务器启动在 http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
