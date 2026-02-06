package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
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
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	
	dbPath := filepath.Join(cfg.DataDir, "subdock.db")
	
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	
	// 自动迁移表结构
	if err := db.AutoMigrate(&Admin{}, &Subscription{}, &Setting{}); err != nil {
		return nil, fmt.Errorf("迁移数据库失败: %w", err)
	}
	
	// 初始化管理员账号
	if err := initAdmin(); err != nil {
		return nil, fmt.Errorf("初始化管理员失败: %w", err)
	}
	
	return db, nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// initAdmin 如果不存在管理员账号，则创建一个
func initAdmin() error {
	var count int64
	if err := db.Model(&Admin{}).Count(&count).Error; err != nil {
		return err
	}
	
	if count > 0 {
		return nil
	}
	
	// 生成随机密码
	password := generateRandomPassword(12)
	
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	admin := &Admin{
		Username:     "admin",
		PasswordHash: string(hash),
	}
	
	if err := db.Create(admin).Error; err != nil {
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

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "subdock123456"
	}
	return hex.EncodeToString(bytes)[:length]
}
