/**
 * 配置数据访问层
 * 提供系统配置的数据库操作
 */
package database

import (
	"database/sql"
	"fmt"
	"strconv"
)

// GetSetting 获取配置值
func GetSetting(key string) (string, error) {
	var value string
	query := `SELECT value FROM settings WHERE key = $1`
	err := DB.QueryRow(query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSetting 设置配置值
func SetSetting(key, value string) error {
	query := `
		INSERT INTO settings (key, value, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := DB.Exec(query, key, value)
	return err
}

// GetMinCodeLength 获取配置的最小代码长度
func GetMinCodeLength() (int, error) {
	value, err := GetSetting("min_code_length")
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil // 使用默认值
	}
	length, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("无效的代码长度配置: %w", err)
	}
	return length, nil
}

// SetMinCodeLength 设置最小代码长度
func SetMinCodeLength(length int) error {
	return SetSetting("min_code_length", strconv.Itoa(length))
}

// GetMaxCodeLength 获取配置的最大代码长度
func GetMaxCodeLength() (int, error) {
	value, err := GetSetting("max_code_length")
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, nil // 使用默认值
	}
	length, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("无效的代码长度配置: %w", err)
	}
	return length, nil
}

// SetMaxCodeLength 设置最大代码长度
func SetMaxCodeLength(length int) error {
	return SetSetting("max_code_length", strconv.Itoa(length))
}

