/**
 * 主程序入口
 * 启动HTTP服务器和初始化所有服务
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

