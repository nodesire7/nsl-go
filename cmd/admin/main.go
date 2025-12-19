/**
 * Admin管理工具
 * 提供命令行工具用于管理admin用户
 */
package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	icfg "short-link/internal/config"
	"short-link/internal/db"
	"short-link/internal/repo"
	"time"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 解析命令行参数
	action := flag.String("action", "", "操作类型: reset-password (重置密码), show-info (显示信息)")
	password := flag.String("password", "", "新密码（可选，不提供则随机生成）")
	flag.Parse()
	
	// 加载配置
	cfg, err := icfg.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	
	// 初始化数据库
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := db.New(ctx, cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer pool.Close()
	
	// 执行迁移
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	
	userRepo := repo.NewUserRepo(pool)
	
	// 执行操作
	switch *action {
	case "reset-password":
		resetAdminPassword(ctx, userRepo, *password)
	case "show-info":
		showAdminInfo(ctx, userRepo)
	case "":
		showUsage()
	default:
		fmt.Printf("未知操作: %s\n", *action)
		showUsage()
		os.Exit(1)
	}
}

// resetAdminPassword 重置admin密码
func resetAdminPassword(ctx context.Context, userRepo *repo.UserRepo, newPassword string) {
	// 检查admin用户是否存在
	admin, err := userRepo.GetAdminUser(ctx)
	if err != nil {
		log.Fatalf("获取admin用户失败: %v", err)
	}
	
	// 如果没有提供密码，生成随机密码
	if newPassword == "" {
		newPassword = generateRandomPassword(16)
	}
	
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码加密失败: %v", err)
	}
	
	// 更新密码
	if err := userRepo.UpdateUserPassword(ctx, "admin", string(hashedPassword)); err != nil {
		log.Fatalf("更新密码失败: %v", err)
	}
	
	fmt.Println("==========================================")
	fmt.Println("✅ Admin密码已更新")
	fmt.Println("==========================================")
	fmt.Printf("用户名: %s\n", admin.Username)
	fmt.Printf("新密码: %s\n", newPassword)
	fmt.Printf("API Token: %s (已改为hash存储，不再显示明文)\n", admin.APIToken)
	fmt.Println("==========================================")
}

// showAdminInfo 显示admin用户信息
func showAdminInfo(ctx context.Context, userRepo *repo.UserRepo) {
	admin, err := userRepo.GetAdminUser(ctx)
	if err != nil {
		log.Fatalf("获取admin用户失败: %v", err)
	}
	
	fmt.Println("==========================================")
	fmt.Println("📋 Admin用户信息")
	fmt.Println("==========================================")
	fmt.Printf("ID: %d\n", admin.ID)
	fmt.Printf("用户名: %s\n", admin.Username)
	fmt.Printf("邮箱: %s\n", admin.Email)
	fmt.Printf("角色: %s\n", admin.Role)
	fmt.Printf("最大链接数: %d (负数表示无限制)\n", admin.MaxLinks)
	fmt.Printf("API Token: %s (已改为hash存储，不再显示明文)\n", admin.APIToken)
	fmt.Printf("创建时间: %s\n", admin.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("==========================================")
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return fmt.Sprintf("admin%d", time.Now().Unix())
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

// showUsage 显示使用说明
func showUsage() {
	fmt.Println("Admin管理工具")
	fmt.Println("")
	fmt.Println("用法:")
	fmt.Println("  nsl-admin -action=reset-password [-password=新密码]")
	fmt.Println("  nsl-admin -action=show-info")
	fmt.Println("")
	fmt.Println("操作说明:")
	fmt.Println("  reset-password  重置admin用户密码（不提供-password参数则随机生成）")
	fmt.Println("  show-info       显示admin用户信息")
	fmt.Println("")
	fmt.Println("示例:")
	fmt.Println("  nsl-admin -action=reset-password")
	fmt.Println("  nsl-admin -action=reset-password -password=MyNewPassword123")
	fmt.Println("  nsl-admin -action=show-info")
}

