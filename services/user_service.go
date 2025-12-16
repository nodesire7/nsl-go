/**
 * 用户服务层
 * 实现用户相关的业务逻辑
 */
package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"short-link/database"
	"short-link/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{}
}

// Register 用户注册
func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {
	// 检查用户名是否存在
	exists, err := database.CheckUsernameExists(req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}
	
	// 检查邮箱是否存在
	exists, err = database.CheckEmailExists(req.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, errors.New("邮箱已被注册")
	}
	
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}
	
	// 生成API Token
	apiToken, err := s.GenerateAPIToken()
	if err != nil {
		return nil, fmt.Errorf("生成API Token失败: %w", err)
	}
	
	// 创建用户（新用户默认最多10条链接）
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		APIToken:  apiToken,
		Role:      "user",
		MaxLinks:  10, // 新用户限制10条
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := database.CreateUser(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	
	return user, nil
}

// Login 用户登录
func (s *UserService) Login(req *models.LoginRequest) (*models.User, error) {
	user, err := database.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	
	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	
	return user, nil
}

// CheckUserLinkLimit 检查用户链接数量限制
func (s *UserService) CheckUserLinkLimit(userID int64) error {
	user, err := database.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}
	
	// -1表示无限制
	if user.MaxLinks == -1 {
		return nil
	}
	
	// 获取用户当前链接数
	count, err := database.GetUserLinkCount(userID)
	if err != nil {
		return fmt.Errorf("获取链接数量失败: %w", err)
	}
	
	if int64(user.MaxLinks) <= count {
		return fmt.Errorf("已达到最大链接数限制（%d条），请联系管理员", user.MaxLinks)
	}
	
	return nil
}

// GetUserInfo 获取用户信息（不包含密码）
func (s *UserService) GetUserInfo(userID int64) (*models.User, error) {
	user, err := database.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	
	// 不返回密码
	user.Password = ""
	return user, nil
}

// GenerateAPIToken 生成API Token
func (s *UserService) GenerateAPIToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "nsl_" + hex.EncodeToString(bytes), nil
}

// UpdateUserToken 更新用户Token
func (s *UserService) UpdateUserToken(userID int64) (string, error) {
	newToken, err := s.GenerateAPIToken()
	if err != nil {
		return "", fmt.Errorf("生成新Token失败: %w", err)
	}
	
	if err := database.UpdateUserToken(userID, newToken); err != nil {
		return "", fmt.Errorf("更新Token失败: %w", err)
	}
	
	return newToken, nil
}

