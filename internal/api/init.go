package api

import (
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"
	"time"

	"golang.org/x/crypto/bcrypt"
	"xorm.io/builder"
)

// 默认管理员账号配置
const DefaultAdminUsername = "admin" // 默认管理员用户名

// InitAdminUser 初始化默认管理员账号
// 首次初始化时使用随机密码，并通过日志输出一次，避免使用固定的弱口令
func InitAdminUser() {
	// 查询是否已存在 admin 账号
	var existing types.User
	found, err := db.FindOne("user", builder.Eq{"username": DefaultAdminUsername}, &existing)
	if err == nil && found && existing.ID != "" {
		// admin 账号已存在，跳过初始化
		return
	}

	// 生成随机初始密码（16 位，crypto/rand 安全随机）
	initialPassword := randomString(16)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(initialPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("管理员默认密码加密失败", "error", err)
		return
	}

	// 创建默认管理员账号
	admin := types.User{
		ID:       db.NewID(),
		Username: DefaultAdminUsername,
		Password: string(hashedPassword),
		Nickname: "管理员",
		Admin:    true,
		Created:  time.Now(),
	}
	if _, err := db.InsertOne("user", admin); err != nil {
		logger.Error("默认管理员账号创建失败", "error", err)
		return
	}
	logger.Info("默认管理员账号已创建，请使用初始密码登录后尽快修改",
		"username", DefaultAdminUsername, "password", initialPassword)
}
