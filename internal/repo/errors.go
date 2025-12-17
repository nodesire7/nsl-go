/**
 * 重写版 Repo 错误定义
 * 统一 repo 层对外返回的错误，便于 handler/service 做稳定判断
 */
package repo

import "errors"

var (
	// ErrNotFound 未找到
	ErrNotFound = errors.New("not found")
)


