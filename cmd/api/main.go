/**
 * 重写版入口
 * 使用 internal/* 分层架构实现
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


