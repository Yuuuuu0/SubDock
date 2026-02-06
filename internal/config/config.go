package config

import (
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	DataDir      string // 数据目录，存放 SQLite 数据库
	Port         int    // HTTP 服务端口
	JWTSecret    string // JWT 签名密钥
	WebsiteTitle string // 网站标题
}

// cfg 全局配置实例
var cfg *Config

// Load 从环境变量加载配置
func Load() *Config {
	cfg = &Config{
		DataDir:      getEnv("DATA_DIR", "./data"),
		Port:         getEnvInt("PORT", 8080),
		JWTSecret:    getEnv("JWT_SECRET", "subdock-default-secret-change-in-production"),
		WebsiteTitle: getEnv("WEBSITE_TITLE", "SubDock"),
	}
	return cfg
}

// Get 获取全局配置
func Get() *Config {
	if cfg == nil {
		return Load()
	}
	return cfg
}

// getEnv 获取环境变量，带默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt 获取整数环境变量，带默认值
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
