package main

import (
	"log"

	"autostack/internal/app"
	"autostack/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化并启动服务器
	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("服务器初始化失败: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
