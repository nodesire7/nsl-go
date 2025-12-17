/**
 * 重写版入口（暂与 cmd/server 共享 internal/app）
 * 后续会把 internal/app 逐步替换为 internal/* 分层架构实现。
 */
package main

import (
	"log"
	"short-link/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}


