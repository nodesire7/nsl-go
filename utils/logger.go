/**
 * 日志工具
 * 提供统一的日志记录功能
 */
package utils

import (
	"log"
	"os"
	"short-link/config"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	WarnLogger  *log.Logger
)

// InitLogger 初始化日志系统
func InitLogger() {
	flags := log.LstdFlags | log.Lshortfile
	
	InfoLogger = log.New(os.Stdout, "[INFO] ", flags)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", flags)
	WarnLogger = log.New(os.Stdout, "[WARN] ", flags)
	
	// 根据配置设置日志级别
	switch config.AppConfig.LogLevel {
	case "DEBUG":
		// 可以添加调试日志
	case "INFO":
		// 默认级别
	case "WARN":
		// 只记录警告和错误
	case "ERROR":
		// 只记录错误
	}
}

// LogInfo 记录信息日志
func LogInfo(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// LogError 记录错误日志
func LogError(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

// LogWarn 记录警告日志
func LogWarn(format string, v ...interface{}) {
	WarnLogger.Printf(format, v...)
}

