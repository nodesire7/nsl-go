/**
 * 链接服务层
 * 实现短链接的核心业务逻辑
 */
package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"short-link/config"
	"short-link/database"
	"short-link/models"
	"short-link/utils"
	"strings"
	"time"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// LinkService 链接服务
type LinkService struct{}

// NewLinkService 创建链接服务实例
func NewLinkService() *LinkService {
	return &LinkService{}
}

// GenerateHash 生成URL的哈希值
func (s *LinkService) GenerateHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomCode 生成随机代码
func (s *LinkService) GenerateRandomCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GetAvailableCodeLength 获取可用的代码长度
func (s *LinkService) GetAvailableCodeLength() (int, error) {
	// 优先使用数据库中的配置
	minLength, err := database.GetMinCodeLength()
	if err != nil || minLength == 0 {
		minLength = config.AppConfig.MinCodeLength
	}
	
	maxLength, err := database.GetMaxCodeLength()
	if err != nil || maxLength == 0 {
		maxLength = config.AppConfig.MaxCodeLength
	}
	
	// 获取当前使用的最大长度
	currentMax, err := database.GetMaxCodeLength()
	if err != nil {
		return minLength, err
	}
	
	// 从最小长度开始检查
	for length := minLength; length <= maxLength; length++ {
		count, err := database.GetCodeCountByLength(length)
		if err != nil {
			return minLength, err
		}
		
		// 计算该长度可以生成的最大数量
		maxPossible := int64(1)
		for i := 0; i < length; i++ {
			maxPossible *= int64(len(charset))
		}
		
		// 如果使用率小于90%，使用该长度
		if float64(count)/float64(maxPossible) < 0.9 {
			return length, nil
		}
	}
	
	// 如果所有长度都用完了，返回最大长度
	return maxLength, nil
}

// CreateLink 创建短链接
func (s *LinkService) CreateLink(userID int64, req *models.CreateLinkRequest, domainService *DomainService) (*models.Link, string, error) {
	// 生成哈希
	hash := s.GenerateHash(req.URL)
	
	// 检查是否已存在相同URL的链接（同一用户）
	existingLink, err := database.GetLinkByHashAndUser(hash, userID)
	if err == nil && existingLink != nil {
		log.Printf("发现已存在的链接，返回现有链接: %s -> %s", req.URL, existingLink.Code)
		// 获取域名URL
		domain, _ := database.GetDomainByID(existingLink.DomainID)
		shortURL := s.buildShortURL(domain, existingLink.Code, domainService)
		return existingLink, shortURL, nil
	}
	
	// 获取域名
	var domainID int64
	if req.DomainID > 0 {
		domainID = req.DomainID
		// 验证域名是否属于该用户
		domain, err := database.GetDomainByID(domainID)
		if err != nil || (domain.UserID != userID && domain.UserID != 0) {
			return nil, "", fmt.Errorf("域名不存在或无权限")
		}
	} else {
		// 获取默认域名
		domain, err := database.GetDefaultDomain(userID)
		if err == nil && domain != nil {
			domainID = domain.ID
		}
	}
	
	// 确定代码
	var code string
	if req.Code != "" {
		// 使用自定义代码
		code = strings.TrimSpace(req.Code)
		// 检查代码是否已存在（同一域名下）
		exists, err := database.CheckCodeExistsInDomain(code, domainID)
		if err != nil {
			return nil, "", fmt.Errorf("检查代码失败: %w", err)
		}
		if exists {
			return nil, "", fmt.Errorf("代码 %s 已存在", code)
		}
	} else {
		// 生成随机代码
		length, err := s.GetAvailableCodeLength()
		if err != nil {
			return nil, "", fmt.Errorf("获取代码长度失败: %w", err)
		}
		
		// 尝试生成唯一代码（最多尝试10次）
		maxAttempts := 10
		for i := 0; i < maxAttempts; i++ {
			code = s.GenerateRandomCode(length)
			exists, err := database.CheckCodeExistsInDomain(code, domainID)
			if err != nil {
				return nil, "", fmt.Errorf("检查代码失败: %w", err)
			}
			if !exists {
				break
			}
			// 如果代码已存在，增加长度重试
			if i == maxAttempts-1 {
				length++
				if length > config.AppConfig.MaxCodeLength {
					return nil, "", fmt.Errorf("无法生成唯一代码，已达到最大长度")
				}
			}
		}
	}
	
	// 获取域名信息用于生成短链接URL
	domain, _ := database.GetDomainByID(domainID)
	shortURL := s.buildShortURL(domain, code, domainService)
	
	// 生成二维码
	qrCode, err := utils.GenerateQRCode(shortURL, 256)
	if err != nil {
		log.Printf("生成二维码失败: %v", err)
		qrCode = "" // 二维码生成失败不影响链接创建
	}
	
	// 创建链接
	link := &models.Link{
		UserID:      userID,
		DomainID:    domainID,
		Code:        code,
		OriginalURL: req.URL,
		Title:       req.Title,
		Hash:        hash,
		QRCode:      qrCode,
		ClickCount:  0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := database.CreateLink(link); err != nil {
		return nil, "", fmt.Errorf("创建链接失败: %w", err)
	}
	
	log.Printf("创建新链接: %s -> %s", req.URL, code)
	return link, shortURL, nil
}

// buildShortURL 构建短链接URL
func (s *LinkService) buildShortURL(domain *models.Domain, code string, domainService *DomainService) string {
	if domain == nil || domain.Domain == "" {
		return config.AppConfig.BaseURL + "/" + code
	}
	
	domainURL := domainService.GetDomainURL(domain)
	if !strings.HasSuffix(domainURL, "/") {
		domainURL += "/"
	}
	return domainURL + code
}

// GetLinkByCode 根据代码获取链接
func (s *LinkService) GetLinkByCode(code string, domainID int64) (*models.Link, error) {
	return database.GetLinkByCode(code, domainID)
}

// GetLinkByCodeAnyDomain 根据代码获取链接（任意域名，用于重定向）
func (s *LinkService) GetLinkByCodeAnyDomain(code string) (*models.Link, error) {
	return database.GetLinkByCodeAnyDomain(code)
}

// RedirectLink 重定向链接（增加点击计数）
func (s *LinkService) RedirectLink(code string, ip, userAgent, referer string) (string, error) {
	link, err := database.GetLinkByCodeAnyDomain(code)
	if err != nil {
		return "", err
	}
	
	// 增加点击计数
	if err := database.IncrementClickCount(link.ID); err != nil {
		log.Printf("增加点击计数失败: %v", err)
	}
	
	// 记录访问日志
	accessLog := &models.AccessLog{
		LinkID:    link.ID,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
		CreatedAt: time.Now(),
	}
	if err := database.CreateAccessLog(accessLog); err != nil {
		log.Printf("记录访问日志失败: %v", err)
	}
	
	return link.OriginalURL, nil
}

// GetLinks 获取链接列表
func (s *LinkService) GetLinks(userID int64, page, limit int) ([]models.Link, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return database.GetUserLinks(userID, page, limit)
}

// DeleteLink 删除链接
func (s *LinkService) DeleteLink(userID int64, code string) error {
	return database.DeleteUserLink(userID, code)
}

// GetStats 获取统计信息
func (s *LinkService) GetStats() (*models.LinkStats, error) {
	return database.GetLinkStats()
}

