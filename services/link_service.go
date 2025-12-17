/**
 * 链接服务层
 * 实现短链接的核心业务逻辑
 */
package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"short-link/cache"
	"short-link/config"
	"short-link/database"
	"short-link/models"
	"short-link/utils"
	"strconv"
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
	// 使用 crypto/rand 生成不可预测随机数，避免 math/rand 可预测/重复 Seed 的安全隐患
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		// 采用拒绝采样避免取模偏差：接受 [0, 248)（62*4）范围
		for {
			var one [1]byte
			if _, err := rand.Read(one[:]); err != nil {
				// 兜底：极少发生，退化为时间哈希（仍然保证可用性）
				h := sha256.Sum256([]byte(fmt.Sprintf("%d:%d", time.Now().UnixNano(), i)))
				b[i] = charset[int(h[0])%len(charset)]
				break
			}
			if one[0] < 248 {
				b[i] = charset[int(one[0])%len(charset)]
				break
			}
		}
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
	// URL 安全校验（SSRF基础防护）
	if err := utils.ValidateExternalURL(req.URL); err != nil {
		return nil, "", fmt.Errorf("URL不合法: %s", err.Error())
	}

	// 生成哈希
	hash := s.GenerateHash(req.URL)
	
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

	// 重写版幂等策略：允许同一URL在不同domain下生成不同短链
	// 因此幂等检查按 (user_id, domain_id, hash) 粒度进行
	if existingLink, err := database.GetLinkByHashUserDomain(hash, userID, domainID); err == nil && existingLink != nil {
		log.Printf("发现已存在的链接（同域名幂等），返回现有链接: %s -> %s", req.URL, existingLink.Code)
		shortURL := s.buildShortURL(domain, existingLink.Code, domainService)
		return existingLink, shortURL, nil
	}

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

	// 并发安全：最终以数据库唯一约束为准，冲突则重试生成
	if req.Code != "" {
		if err := database.CreateLink(link); err != nil {
			if database.IsUniqueViolation(err) {
				return nil, "", fmt.Errorf("代码 %s 已存在", code)
			}
			return nil, "", fmt.Errorf("创建链接失败: %w", err)
		}
	} else {
		createAttempts := 0
		for {
			createAttempts++
			err := database.CreateLink(link)
			if err == nil {
				break
			}
			if database.IsUniqueViolation(err) && createAttempts < 20 {
				// 冲突：增加长度或重试
				if createAttempts%10 == 0 {
					// 每10次冲突增加长度（受最大值限制）
					nextLen := len(link.Code) + 1
					if nextLen > config.AppConfig.MaxCodeLength {
						return nil, "", fmt.Errorf("无法生成唯一代码，已达到最大长度")
					}
					link.Code = s.GenerateRandomCode(nextLen)
				} else {
					link.Code = s.GenerateRandomCode(len(link.Code))
				}
				// 更新短链接与二维码（保持一致）
				code = link.Code
				shortURL = s.buildShortURL(domain, code, domainService)
				qrCode, _ := utils.GenerateQRCode(shortURL, 256)
				link.QRCode = qrCode
				continue
			}
			return nil, "", fmt.Errorf("创建链接失败: %w", err)
		}
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
	// 热点缓存：优先从Redis读取（仅缓存读取路径，写入仍落DB）
	if cache.RedisClient != nil {
		if v, err := cache.Get("redir:" + code); err == nil && v != "" {
			// 格式：<id>|<url>
			parts := strings.SplitN(v, "|", 2)
			if len(parts) == 2 {
				if id, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					// 增加点击计数（DB写）
					if err := database.IncrementClickCount(id); err != nil {
						log.Printf("增加点击计数失败: %v", err)
					}
					// 记录访问日志（DB写）
					accessLog := &models.AccessLog{
						LinkID:    id,
						IP:        ip,
						UserAgent: userAgent,
						Referer:   referer,
						CreatedAt: time.Now(),
					}
					if err := database.CreateAccessLog(accessLog); err != nil {
						log.Printf("记录访问日志失败: %v", err)
					}
					return parts[1], nil
				}
			}
		}
	}

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

	// 写入缓存（1小时）
	if cache.RedisClient != nil {
		_ = cache.Set("redir:"+code, fmt.Sprintf("%d|%s", link.ID, link.OriginalURL), time.Hour)
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
	if err := database.DeleteUserLink(userID, code); err != nil {
		return err
	}
	// 删除缓存（best-effort）
	_ = cache.Delete("redir:" + code)
	return nil
}

// GetStats 获取统计信息
func (s *LinkService) GetStats() (*models.LinkStats, error) {
	return database.GetLinkStats()
}

