/**
 * 域名数据访问层
 * 提供域名的数据库操作
 */
package database

import (
	"database/sql"
	"fmt"
	"short-link/models"
	"time"
)

// CreateDomain 创建域名
func CreateDomain(domain *models.Domain) error {
	query := `
		INSERT INTO domains (user_id, domain, is_default, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	
	err := DB.QueryRow(
		query,
		domain.UserID,
		domain.Domain,
		domain.IsDefault,
		domain.IsActive,
		domain.CreatedAt,
		domain.UpdatedAt,
	).Scan(&domain.ID)
	
	return err
}

// GetDomainByID 根据ID获取域名
func GetDomainByID(domainID int64) (*models.Domain, error) {
	domain := &models.Domain{}
	query := `SELECT id, user_id, domain, is_default, is_active, created_at, updated_at 
			  FROM domains WHERE id = $1`
	
	err := DB.QueryRow(query, domainID).Scan(
		&domain.ID,
		&domain.UserID,
		&domain.Domain,
		&domain.IsDefault,
		&domain.IsActive,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("域名不存在")
	}
	
	return domain, err
}

// GetUserDomains 获取用户的所有域名
func GetUserDomains(userID int64) ([]models.Domain, error) {
	query := `SELECT id, user_id, domain, is_default, is_active, created_at, updated_at 
			  FROM domains WHERE user_id = $1 OR user_id = 0
			  ORDER BY is_default DESC, created_at ASC`
	
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var domains []models.Domain
	for rows.Next() {
		var domain models.Domain
		err := rows.Scan(
			&domain.ID,
			&domain.UserID,
			&domain.Domain,
			&domain.IsDefault,
			&domain.IsActive,
			&domain.CreatedAt,
			&domain.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	
	return domains, nil
}

// GetDefaultDomain 获取默认域名
func GetDefaultDomain(userID int64) (*models.Domain, error) {
	// 先查找用户的默认域名
	domain := &models.Domain{}
	query := `SELECT id, user_id, domain, is_default, is_active, created_at, updated_at 
			  FROM domains WHERE user_id = $1 AND is_default = true AND is_active = true
			  LIMIT 1`
	
	err := DB.QueryRow(query, userID).Scan(
		&domain.ID,
		&domain.UserID,
		&domain.Domain,
		&domain.IsDefault,
		&domain.IsActive,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)
	
	if err == nil {
		return domain, nil
	}
	
	// 如果没有找到，查找系统默认域名
	query = `SELECT id, user_id, domain, is_default, is_active, created_at, updated_at 
			 FROM domains WHERE user_id = 0 AND is_default = true AND is_active = true
			 LIMIT 1`
	
	err = DB.QueryRow(query).Scan(
		&domain.ID,
		&domain.UserID,
		&domain.Domain,
		&domain.IsDefault,
		&domain.IsActive,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		// 如果都没有，返回系统配置的BASE_URL
		return &models.Domain{
			ID:        0,
			UserID:    0,
			Domain:    "", // 使用系统配置
			IsDefault: true,
			IsActive:  true,
		}, nil
	}
	
	return domain, err
}

// DeleteDomain 删除域名
func DeleteDomain(domainID, userID int64) error {
	query := `DELETE FROM domains WHERE id = $1 AND user_id = $2`
	result, err := DB.Exec(query, domainID, userID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("域名不存在或无权限删除")
	}
	
	return nil
}

// SetDefaultDomain 设置默认域名
func SetDefaultDomain(domainID, userID int64) error {
	// 先取消该用户的所有默认域名
	query := `UPDATE domains SET is_default = false WHERE user_id = $1`
	_, err := DB.Exec(query, userID)
	if err != nil {
		return err
	}
	
	// 设置新的默认域名
	query = `UPDATE domains SET is_default = true, updated_at = $1 WHERE id = $2 AND user_id = $3`
	_, err = DB.Exec(query, time.Now(), domainID, userID)
	return err
}

// CheckDomainExists 检查域名是否存在
func CheckDomainExists(domain string, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM domains WHERE domain = $1 AND (user_id = $2 OR user_id = 0))`
	err := DB.QueryRow(query, domain, userID).Scan(&exists)
	return exists, err
}

