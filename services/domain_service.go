/**
 * 域名服务层
 * 实现域名相关的业务逻辑
 */
package services

import (
	"fmt"
	"short-link/config"
	"short-link/database"
	"short-link/models"
	"strings"
	"time"
)

// DomainService 域名服务
type DomainService struct{}

// NewDomainService 创建域名服务实例
func NewDomainService() *DomainService {
	return &DomainService{}
}

// CreateDomain 创建域名
func (s *DomainService) CreateDomain(userID int64, req *models.CreateDomainRequest) (*models.Domain, error) {
	// 验证域名格式
	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		return nil, fmt.Errorf("域名不能为空")
	}
	
	// 移除协议前缀
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "//")
	
	// 检查域名是否已存在
	exists, err := database.CheckDomainExists(domain, userID)
	if err != nil {
		return nil, fmt.Errorf("检查域名失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("域名已存在")
	}
	
	// 如果设置为默认，先取消其他默认域名
	if req.IsDefault {
		domains, _ := database.GetUserDomains(userID)
		for _, d := range domains {
			if d.IsDefault && d.UserID == userID {
				_ = database.SetDefaultDomain(0, userID) // 先取消
			}
		}
	}
	
	domainModel := &models.Domain{
		UserID:    userID,
		Domain:    domain,
		IsDefault: req.IsDefault,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := database.CreateDomain(domainModel); err != nil {
		return nil, fmt.Errorf("创建域名失败: %w", err)
	}
	
	return domainModel, nil
}

// GetUserDomains 获取用户的所有域名
func (s *DomainService) GetUserDomains(userID int64) ([]models.Domain, error) {
	return database.GetUserDomains(userID)
}

// GetDefaultDomain 获取默认域名
func (s *DomainService) GetDefaultDomain(userID int64) (*models.Domain, error) {
	domain, err := database.GetDefaultDomain(userID)
	if err != nil {
		return nil, err
	}
	
	// 如果域名为空，使用系统配置的BASE_URL
	if domain.Domain == "" {
		baseURL := config.AppConfig.BaseURL
		// 从BASE_URL中提取域名
		domain.Domain = strings.TrimPrefix(baseURL, "http://")
		domain.Domain = strings.TrimPrefix(domain.Domain, "https://")
	}
	
	return domain, nil
}

// DeleteDomain 删除域名
func (s *DomainService) DeleteDomain(domainID, userID int64) error {
	return database.DeleteDomain(domainID, userID)
}

// SetDefaultDomain 设置默认域名
func (s *DomainService) SetDefaultDomain(domainID, userID int64) error {
	return database.SetDefaultDomain(domainID, userID)
}

// GetDomainURL 获取完整的域名URL（带协议）
func (s *DomainService) GetDomainURL(domain *models.Domain) string {
	if domain.Domain == "" {
		return config.AppConfig.BaseURL
	}
	
	// 如果域名不包含协议，添加https://
	if !strings.HasPrefix(domain.Domain, "http://") && !strings.HasPrefix(domain.Domain, "https://") {
		return "https://" + domain.Domain
	}
	
	return domain.Domain
}

