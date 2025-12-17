/**
 * URL 校验（基础 SSRF 防护）
 * 默认策略：
 * - 仅允许 http/https
 * - 禁止 localhost / 127.0.0.1 / ::1 / 私有网段 / link-local / multicast
 * - 对域名进行DNS解析并校验解析结果（可通过环境变量关闭）
 *
 * 可通过环境变量调整：
 * - ALLOW_PRIVATE_URLS=true 允许内网地址（默认 false）
 * - URL_VALIDATE_DNS=false 关闭DNS解析校验（默认 true）
 */
package utils

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"time"
)

// ValidateExternalURL 校验外部URL是否合法且不指向内网
func ValidateExternalURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return errors.New("URL不能为空")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("URL格式不正确")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("仅允许 http/https URL")
	}
	if u.Host == "" {
		return errors.New("URL缺少host")
	}
	if u.User != nil {
		return errors.New("URL不允许包含用户名密码")
	}

	host := u.Hostname()
	if host == "" {
		return errors.New("URL host无效")
	}

	if allowPrivateURLs() {
		return nil
	}

	// 直接 IP 检查
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return errors.New("不允许内网/本机地址")
		}
		return nil
	}

	// 常见 localhost 域名
	lh := strings.ToLower(host)
	if lh == "localhost" || strings.HasSuffix(lh, ".localhost") || strings.HasSuffix(lh, ".local") {
		return errors.New("不允许内网/本机地址")
	}

	// DNS 解析校验
	if !validateDNS() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		// DNS失败：保守起见拒绝（避免绕过）
		return errors.New("无法解析URL域名")
	}
	for _, a := range addrs {
		if isPrivateIP(a.IP) {
			return errors.New("不允许内网/本机地址")
		}
	}
	return nil
}

func allowPrivateURLs() bool {
	return strings.EqualFold(os.Getenv("ALLOW_PRIVATE_URLS"), "true")
}

func validateDNS() bool {
	v := os.Getenv("URL_VALIDATE_DNS")
	if v == "" {
		return true
	}
	return !strings.EqualFold(v, "false")
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return true
	}
	// Go 1.21: IsPrivate 覆盖 RFC1918/4193
	if ip.IsPrivate() {
		return true
	}
	// 0.0.0.0/8, ::/128 等
	if ip.IsUnspecified() {
		return true
	}
	return false
}


