package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"subdock/internal/config"
)

var db *gorm.DB

// InitDB 初始化数据库连接和表结构
func InitDB() (*gorm.DB, error) {
	cfg := config.Get()

	// 确保数据目录存在
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dbPath := filepath.Join(cfg.DataDir, "subdock.db")
	firstRun := isFirstRun(dbPath)
	absDBPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("解析数据库路径失败: %w", err)
	}
	dsn := (&url.URL{Scheme: "file", Path: absDBPath}).String() + "?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on"

	db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}
	// 单文件 SQLite 在单进程内串行写入，配合条件更新避免调度器与 HTTP 请求并发丢更新。
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if firstRun {
		log.Printf("检测到首次启动，数据库文件不存在，将初始化新库: %s", dbPath)
	} else {
		log.Printf("检测到已存在数据库文件，将执行结构一致性检查: %s", dbPath)
	}

	// 启动时执行结构校验与迁移：缺表创建、缺字段补齐、结构按模型同步
	if err := ensureSchema(db); err != nil {
		return nil, fmt.Errorf("同步数据库结构失败: %w", err)
	}

	// 初始化管理员账号
	if err := initAdmin(db); err != nil {
		return nil, fmt.Errorf("初始化管理员失败: %w", err)
	}

	return db, nil
}

// isFirstRun 判断是否首次运行（数据库文件不存在）
func isFirstRun(dbPath string) bool {
	_, err := os.Stat(dbPath)
	return os.IsNotExist(err)
}

// ensureSchema 确保数据库结构与模型一致
func ensureSchema(db *gorm.DB) error {
	models := []interface{}{&Admin{}, &Subscription{}, &SubscriptionRenewal{}, &Setting{}}
	migrator := db.Migrator()

	// 1) 缺表时创建
	for _, m := range models {
		if !migrator.HasTable(m) {
			if err := migrator.CreateTable(m); err != nil {
				return fmt.Errorf("创建表失败: %w", err)
			}
		}
	}

	// 2) 关键字段兜底（历史库升级场景）
	if !migrator.HasColumn(&Subscription{}, "auto_renew") {
		if err := migrator.AddColumn(&Subscription{}, "AutoRenew"); err != nil {
			return fmt.Errorf("补齐 subscriptions.auto_renew 失败: %w", err)
		}
	}
	if !migrator.HasColumn(&Subscription{}, "renew_count") {
		if err := migrator.AddColumn(&Subscription{}, "RenewCount"); err != nil {
			return fmt.Errorf("补齐 subscriptions.renew_count 失败: %w", err)
		}
	}

	// 3) 按模型执行自动迁移（类型/索引等结构同步）
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// initAdmin 如果不存在管理员账号，则创建一个并输出一次性初始密码。
func initAdmin(database *gorm.DB) error {
	var count int64
	if err := database.Model(&Admin{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	// 生成随机密码
	password, err := generateRandomPassword(24)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &Admin{
		Username:     "admin",
		PasswordHash: string(hash),
	}

	if err := database.Create(admin).Error; err != nil {
		return err
	}

	// 打印初始密码到日志
	log.Printf("========================================")
	log.Printf("初始管理员账号已创建")
	log.Printf("用户名: admin")
	log.Printf("密码: %s", password)
	log.Printf("请登录后立即修改密码！")
	log.Printf("========================================")

	return nil
}

// generateRandomPassword 生成指定长度的十六进制随机密码，随机源失败时返回错误而非弱默认值。
func generateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("密码长度必须大于 0")
	}
	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成管理员随机密码失败: %w", err)
	}
	return hex.EncodeToString(bytes)[:length], nil
}
