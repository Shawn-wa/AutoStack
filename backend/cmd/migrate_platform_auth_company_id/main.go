package main

import (
	"fmt"
	"log"

	"autostack/internal/commonBase/database"
	"autostack/internal/config"
	"autostack/internal/migration/companyid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	db := database.GetDB()

	if err := companyid.Run(db); err != nil {
		log.Fatalf("company_id 迁移失败: %v", err)
	}

	fmt.Println("全业务表 company_id 迁移完成")
}
