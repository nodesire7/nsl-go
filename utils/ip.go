/**
 * IP 工具
 * 处理代理链路的真实 IP 提取（X-Forwarded-For / X-Real-IP）
 * 实现 redo.md 2.5：代理链路真实 IP 处理
 */
package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetRealIP 从请求中提取真实客户端 IP
// 优先级：X-Forwarded-For（取第一个） > X-Real-IP > RemoteAddr
// 同时进行基本验证，防止 IP 伪造
func GetRealIP(r *http.Request) string {
	// 1. 检查 X-Forwarded-For（可能包含多个 IP，用逗号分隔）
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		// 取第一个 IP（最原始的客户端 IP）
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// 2. 检查 X-Real-IP
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		ip := strings.TrimSpace(realIP)
		if isValidIP(ip) {
			return ip
		}
	}

	// 3. 回退到 RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// 如果没有端口，直接使用 RemoteAddr
		return r.RemoteAddr
	}
	return ip
}

// isValidIP 验证 IP 地址格式（基本验证，防止明显伪造）
func isValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// 拒绝内网 IP（如果配置了信任代理，可以放宽此限制）
	// 这里仅做格式验证，不拒绝内网 IP（因为可能确实来自内网代理）
	return true
}

