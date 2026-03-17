package main

import (
	"flag"
	"fmt"
	"log"

	"autostack/internal/commonBase/database"
	"autostack/internal/config"
	"autostack/internal/modules/user"
)

func main() {
	rebuildRoleBindings := flag.Bool("rebuild-role-bindings", false, "是否补齐所有企业默认角色权限绑定")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	db := database.GetDB()
	user.InitHandler(db)

	result, err := user.GetService().RunPermissionMigration(*rebuildRoleBindings)
	if err != nil {
		log.Fatalf("权限初始化迁移失败: %v", err)
	}

	fmt.Printf("权限初始化迁移完成，企业总数=%d，已处理=%d，重建角色绑定=%t\n",
		result.CompaniesTotal, result.CompaniesProcessed, result.RebuildRoleBinding)
}
