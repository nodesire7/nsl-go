/**
 * 数据库错误辅助
 * 用于识别常见错误类型（如唯一约束冲突），便于服务层做重试与更友好的错误提示。
 */
package database

import "github.com/lib/pq"

// IsUniqueViolation 判断是否为唯一约束冲突（PostgreSQL 23505）
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return string(pqErr.Code) == "23505"
	}
	return false
}


