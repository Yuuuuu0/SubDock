package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const jwtSecretFileName = ".jwt-secret"

// Config 表示应用运行所需的环境配置。
type Config struct {
	DataDir      string // 数据目录，存放 SQLite 数据库和运行密钥
	Port         int    // HTTP 服务端口
	JWTSecret    string // JWT 签名密钥
	WebsiteTitle string // 网站标题
}

var cfg *Config

// Load 从环境变量加载并校验配置，未配置 JWT_SECRET 时会在数据目录持久化安全随机密钥。
func Load() (*Config, error) {
	dataDir := getEnv("DATA_DIR", "./data")
	jwtSecret, err := loadJWTSecret(dataDir)
	if err != nil {
		return nil, err
	}

	port, err := getEnvInt("PORT", 8080)
	if err != nil {
		return nil, err
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("PORT 必须在 1 到 65535 之间")
	}

	cfg = &Config{
		DataDir:      dataDir,
		Port:         port,
		JWTSecret:    jwtSecret,
		WebsiteTitle: getEnv("WEBSITE_TITLE", "SubDock"),
	}
	return cfg, nil
}

// Get 获取已加载的全局配置；应用入口必须先调用 Load。
func Get() *Config {
	if cfg == nil {
		panic("配置尚未加载")
	}
	return cfg
}

// loadJWTSecret 优先读取环境变量，否则读取或创建数据目录中的持久化密钥文件。
func loadJWTSecret(dataDir string) (string, error) {
	if value, ok := os.LookupEnv("JWT_SECRET"); ok && strings.TrimSpace(value) != "" {
		secret := strings.TrimSpace(value)
		if err := validateJWTSecret(secret); err != nil {
			return "", fmt.Errorf("JWT_SECRET 不安全: %w", err)
		}
		return secret, nil
	}

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return "", fmt.Errorf("创建数据目录失败: %w", err)
	}

	secretPath := filepath.Join(dataDir, jwtSecretFileName)
	secret, err := readJWTSecretFile(secretPath)
	if err == nil {
		return secret, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	secret, err = generateJWTSecret()
	if err != nil {
		return "", err
	}
	file, err := os.OpenFile(secretPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return readJWTSecretFile(secretPath)
		}
		return "", fmt.Errorf("创建 JWT 密钥文件失败: %w", err)
	}

	writeErr := func() error {
		defer file.Close()
		if _, err := file.WriteString(secret + "\n"); err != nil {
			return fmt.Errorf("写入 JWT 密钥文件失败: %w", err)
		}
		if err := file.Sync(); err != nil {
			return fmt.Errorf("同步 JWT 密钥文件失败: %w", err)
		}
		return nil
	}()
	if writeErr != nil {
		_ = os.Remove(secretPath)
		return "", writeErr
	}
	return secret, nil
}

// readJWTSecretFile 读取并校验持久化密钥，同时确保文件权限仅限当前用户。
func readJWTSecretFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", os.ErrNotExist
		}
		return "", fmt.Errorf("读取 JWT 密钥文件失败: %w", err)
	}
	if err := os.Chmod(path, 0600); err != nil {
		return "", fmt.Errorf("设置 JWT 密钥文件权限失败: %w", err)
	}
	secret := strings.TrimSpace(string(content))
	if err := validateJWTSecret(secret); err != nil {
		return "", fmt.Errorf("JWT 密钥文件无效: %w", err)
	}
	return secret, nil
}

// generateJWTSecret 生成 256 位随机 JWT 密钥并以十六进制编码返回。
func generateJWTSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成 JWT 随机密钥失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// validateJWTSecret 拒绝长度不足和常见示例占位密钥。
func validateJWTSecret(secret string) error {
	weakValues := map[string]struct{}{
		"subdock-default-secret-change-in-production": {},
		"your_jwt_secret": {},
		"change-me":       {},
		"changeme":        {},
		"secret":          {},
	}
	if _, weak := weakValues[strings.ToLower(secret)]; weak {
		return errors.New("不能使用默认或示例密钥")
	}
	if len(secret) < 32 {
		return errors.New("长度至少需要 32 个字符")
	}
	return nil
}

// getEnv 获取环境变量，未设置时返回默认值。
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt 获取整数环境变量，格式无效时返回明确错误。
func getEnvInt(key string, defaultVal int) (int, error) {
	if val := os.Getenv(key); val != "" {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("%s 必须是整数: %w", key, err)
		}
		return intVal, nil
	}
	return defaultVal, nil
}
