package model

import (
	"os"
	"path/filepath"
	"testing"

	"subdock/internal/config"
)

// TestInitDBCreatesPrivateDataDirectory 验证数据库目录默认仅允许当前用户访问。
func TestInitDBCreatesPrivateDataDirectory(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "private-data")
	t.Setenv("DATA_DIR", dataDir)
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	if _, err := config.Load(); err != nil {
		t.Fatalf("加载测试配置失败: %v", err)
	}
	database, err := InitDB()
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("获取测试数据库连接失败: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	info, err := os.Stat(dataDir)
	if err != nil {
		t.Fatalf("读取数据目录信息失败: %v", err)
	}
	if permission := info.Mode().Perm(); permission != 0700 {
		t.Fatalf("数据目录权限为 %o，期望 700", permission)
	}
}
