/**
 * 用户 Service（重写版）
 * - 业务逻辑与 repo 解耦
 * - 统一处理密码 hash、token 生成、用户限制等
 */
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"short-link/internal/repo"
	"short-link/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务（重写版）
type UserService struct {
	userRepo *repo.UserRepo
}

// NewUserService 创建 UserService
func NewUserService(userRepo *repo.UserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

// Register 注册
func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	exists, err := s.userRepo.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	exists, err = s.userRepo.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("检查邮箱失败: %w", err)
	}
	if exists {
		return nil, errors.New("邮箱已被注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	apiToken, err := GenerateAPIToken()
	if err != nil {
		return nil, fmt.Errorf("生成API Token失败: %w", err)
	}

	now := time.Now()
	u := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		APIToken:  apiToken,
		Role:      "user",
		MaxLinks:  10,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.userRepo.CreateUser(ctx, u); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return u, nil
}

// Login 登录
func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.User, error) {
	u, err := s.userRepo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return u, nil
}

// GetUserInfo 获取用户信息（不包含密码）
func (s *UserService) GetUserInfo(ctx context.Context, userID int64) (*models.User, error) {
	u, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

// RotateAPIToken 轮换用户 API Token
func (s *UserService) RotateAPIToken(ctx context.Context, userID int64) (string, error) {
	newToken, err := GenerateAPIToken()
	if err != nil {
		return "", fmt.Errorf("生成新Token失败: %w", err)
	}
	if err := s.userRepo.UpdateUserToken(ctx, userID, newToken); err != nil {
		return "", fmt.Errorf("更新Token失败: %w", err)
	}
	return newToken, nil
}

// GenerateAPIToken 生成 API Token（永久有效）
func GenerateAPIToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "nsl_" + hex.EncodeToString(b), nil
}


