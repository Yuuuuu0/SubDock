package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadGeneratesPersistentJWTSecret 验证未配置密钥时会生成可复用的 0600 文件。
func TestLoadGeneratesPersistentJWTSecret(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("DATA_DIR", dataDir)
	t.Setenv("JWT_SECRET", "")
	t.Setenv("PORT", "")

	first, err := Load()
	if err != nil {
		t.Fatalf("首次加载配置失败: %v", err)
	}
	if len(first.JWTSecret) < 32 {
		t.Fatalf("生成的 JWT 密钥过短: %d", len(first.JWTSecret))
	}

	second, err := Load()
	if err != nil {
		t.Fatalf("再次加载配置失败: %v", err)
	}
	if second.JWTSecret != first.JWTSecret {
		t.Fatal("JWT 密钥未在重启语义下保持稳定")
	}

	info, err := os.Stat(filepath.Join(dataDir, jwtSecretFileName))
	if err != nil {
		t.Fatalf("读取 JWT 密钥文件信息失败: %v", err)
	}
	if permission := info.Mode().Perm(); permission != 0600 {
		t.Fatalf("JWT 密钥文件权限为 %o，期望 600", permission)
	}
}

// TestLoadRejectsWeakJWTSecret 验证常见部署占位值不会被接受。
func TestLoadRejectsWeakJWTSecret(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("JWT_SECRET", "your_jwt_secret")

	if _, err := Load(); err == nil {
		t.Fatal("弱 JWT_SECRET 应被拒绝")
	}
}

// TestLoadRejectsInvalidPort 验证端口格式和范围错误会阻止启动。
func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("PORT", "70000")

	if _, err := Load(); err == nil {
		t.Fatal("越界 PORT 应被拒绝")
	}
}
