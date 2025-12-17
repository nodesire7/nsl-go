/**
 * Link Service（重写版）
 * - 短链创建：幂等（user+domain+hash）、动态长度、并发安全重试
 * - URL 校验：复用 utils.ValidateExternalURL（基础 SSRF 防护）
 */
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"short-link/cache"
	"short-link/internal/repo"
	"short-link/models"
	"short-link/utils"
	"strings"
	"time"
)

const v2Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// LinkService 链接服务（重写版）
type LinkService struct {
	linkRepo     *repo.LinkRepo
	domainRepo   *repo.DomainRepo
	settingsRepo *repo.SettingsRepo
	userRepo     *repo.UserRepo
	accessLogRepo *repo.AccessLogRepo

	// env 默认值（DB settings 可覆盖）
	minCodeLen int
	maxCodeLen int
	baseURL    string
}

// NewLinkService 创建 LinkService
func NewLinkService(baseURL string, minCodeLen int, maxCodeLen int, linkRepo *repo.LinkRepo, domainRepo *repo.DomainRepo, settingsRepo *repo.SettingsRepo, userRepo *repo.UserRepo, accessLogRepo *repo.AccessLogRepo) *LinkService {
	return &LinkService{
		linkRepo:     linkRepo,
		domainRepo:   domainRepo,
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
		accessLogRepo: accessLogRepo,
		minCodeLen:   minCodeLen,
		maxCodeLen:   maxCodeLen,
		baseURL:      baseURL,
	}
}

// GenerateHash 生成 URL 内容 hash（SHA256 hex）
func (s *LinkService) GenerateHash(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])
}

// GenerateRandomCode 生成随机短码（crypto/rand + 拒绝采样）
func (s *LinkService) GenerateRandomCode(length int) string {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		for {
			var one [1]byte
			if _, err := rand.Read(one[:]); err != nil {
				h := sha256.Sum256([]byte(fmt.Sprintf("%d:%d", time.Now().UnixNano(), i)))
				b[i] = v2Charset[int(h[0])%len(v2Charset)]
				break
			}
			// 62*4=248，接受 [0,248) 避免取模偏差
			if one[0] < 248 {
				b[i] = v2Charset[int(one[0])%len(v2Charset)]
				break
			}
		}
	}
	return string(b)
}

// getMinMaxCodeLength 获取 min/max（settings 优先，env 兜底）
func (s *LinkService) getMinMaxCodeLength(ctx context.Context) (int, int, error) {
	min := s.minCodeLen
	max := s.maxCodeLen

	if s.settingsRepo != nil {
		if v, err := s.settingsRepo.GetMinCodeLength(ctx); err == nil && v > 0 {
			min = v
		}
		if v, err := s.settingsRepo.GetMaxCodeLength(ctx); err == nil && v > 0 {
			max = v
		}
	}

	if min <= 0 || max <= 0 || min > max {
		return 0, 0, fmt.Errorf("短码长度配置无效（min=%d max=%d）", min, max)
	}
	return min, max, nil
}

// maxPossibleCount 计算给定长度的最大可用数量（62^len）
func maxPossibleCount(length int) float64 {
	m := 1.0
	for i := 0; i < length; i++ {
		m *= float64(len(v2Charset))
	}
	return m
}

// GetAvailableCodeLength 获取可用短码长度：从 min 开始，若已穷尽则递增
func (s *LinkService) GetAvailableCodeLength(ctx context.Context) (int, error) {
	min, max, err := s.getMinMaxCodeLength(ctx)
	if err != nil {
		return 0, err
	}

	for l := min; l <= max; l++ {
		count, err := s.linkRepo.GetCodeCountByLength(ctx, l)
		if err != nil {
			return min, err
		}
		if float64(count) < maxPossibleCount(l) {
			return l, nil
		}
	}
	return max, nil
}

// BuildShortURL 生成完整短链接
func (s *LinkService) BuildShortURL(domain *models.Domain, code string) string {
	base := ""
	if domain != nil && strings.TrimSpace(domain.Domain) != "" {
		d := strings.TrimSpace(domain.Domain)
		if strings.HasPrefix(d, "http://") || strings.HasPrefix(d, "https://") {
			base = d
		} else {
			base = "https://" + d
		}
	} else {
		base = s.baseURL
	}
	base = strings.TrimRight(base, "/")
	return base + "/" + code
}

func normalizeHost(hostport string) (string, string) {
	hostport = strings.TrimSpace(strings.ToLower(hostport))
	if hostport == "" {
		return "", ""
	}
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil && h != "" {
		host = h
	}
	return hostport, host
}

func baseURLHosts(baseURL string) (string, string) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" {
		return normalizeHost(baseURL)
	}
	return normalizeHost(u.Host)
}

// ResolveDomainForHost 根据请求 Host 解析 domain 记录
// - 若 host 匹配 BaseURL 的 host，则返回系统默认域名（user_id=0, is_default=true）
// - 否则按 domains.domain 精确匹配（允许带/不带端口）
func (s *LinkService) ResolveDomainForHost(ctx context.Context, hostport string) (*models.Domain, error) {
	if s.domainRepo == nil {
		return nil, repo.ErrNotFound
	}

	reqHostport, reqHost := normalizeHost(hostport)
	baseHostport, baseHost := baseURLHosts(s.baseURL)

	// BaseURL Host 命中：系统默认域名
	if reqHostport != "" && (reqHostport == baseHostport || reqHost == baseHost) {
		return s.domainRepo.GetDefaultDomain(ctx, 0)
	}

	// 自定义域名：先尝试带端口/不带端口
	candidates := map[string]struct{}{}
	if reqHostport != "" {
		candidates[reqHostport] = struct{}{}
	}
	if reqHost != "" {
		candidates[reqHost] = struct{}{}
	}

	var found []models.Domain
	for name := range candidates {
		ds, err := s.domainRepo.FindActiveDomainsByName(ctx, name)
		if err != nil {
			return nil, err
		}
		found = append(found, ds...)
	}

	if len(found) == 0 {
		return nil, repo.ErrNotFound
	}
	if len(found) > 1 {
		// 域名配置冲突：同一 host 对应多条记录（需要管理员清理/加唯一约束）
		return nil, fmt.Errorf("域名配置冲突：%s 对应多条记录", reqHostport)
	}
	return &found[0], nil
}

// RedirectLink v2 重定向解析（含热点缓存 + 点击/日志写入）
func (s *LinkService) RedirectLink(ctx context.Context, hostport string, code string, ip string, userAgent string, referer string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", repo.ErrNotFound
	}

	// 解析域名（优先）
	domain, _ := s.ResolveDomainForHost(ctx, hostport)
	domainID := int64(0)
	if domain != nil {
		domainID = domain.ID
	}

	cacheKey := fmt.Sprintf("redir:%d:%s", domainID, code)
	if cache.RedisClient != nil {
		if v, err := cache.Get(cacheKey); err == nil && v != "" {
			parts := strings.SplitN(v, "|", 2)
			if len(parts) == 2 {
				// best-effort 写入统计
				_ = s.linkRepo.IncrementClickCount(ctx, parseInt64(parts[0]))
				if s.accessLogRepo != nil {
					_ = s.accessLogRepo.CreateAccessLog(ctx, &models.AccessLog{
						LinkID:    parseInt64(parts[0]),
						IP:        ip,
						UserAgent: userAgent,
						Referer:   referer,
						CreatedAt: time.Now(),
					})
				}
				return parts[1], nil
			}
		}
	}

	// 主查：domain + code
	if domain != nil {
		l, err := s.linkRepo.GetLinkByCode(ctx, code, domain.ID)
		if err == nil {
			_ = s.linkRepo.IncrementClickCount(ctx, l.ID)
			if s.accessLogRepo != nil {
				_ = s.accessLogRepo.CreateAccessLog(ctx, &models.AccessLog{
					LinkID:    l.ID,
					IP:        ip,
					UserAgent: userAgent,
					Referer:   referer,
					CreatedAt: time.Now(),
				})
			}
			if cache.RedisClient != nil {
				_ = cache.Set(cacheKey, fmt.Sprintf("%d|%s", l.ID, l.OriginalURL), time.Hour)
			}
			return l.OriginalURL, nil
		}
	}

	// 兼容回退：host 未识别时，若全库只有一个 code 命中则允许跳转，否则 404
	ls, err := s.linkRepo.GetLinkByCodeAnyDomain(ctx, code, 2)
	if err != nil {
		return "", err
	}
	if len(ls) != 1 {
		return "", repo.ErrNotFound
	}
	l := ls[0]
	_ = s.linkRepo.IncrementClickCount(ctx, l.ID)
	if s.accessLogRepo != nil {
		_ = s.accessLogRepo.CreateAccessLog(ctx, &models.AccessLog{
			LinkID:    l.ID,
			IP:        ip,
			UserAgent: userAgent,
			Referer:   referer,
			CreatedAt: time.Now(),
		})
	}
	if cache.RedisClient != nil {
		_ = cache.Set(fmt.Sprintf("redir:%d:%s", l.DomainID, code), fmt.Sprintf("%d|%s", l.ID, l.OriginalURL), time.Hour)
	}
	return l.OriginalURL, nil
}

func parseInt64(s string) int64 {
	var n int64
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int64(ch-'0')
	}
	return n
}

// CreateLink 创建短链（v2）
func (s *LinkService) CreateLink(ctx context.Context, userID int64, req *models.CreateLinkRequest) (*models.Link, string, error) {
	if err := utils.ValidateExternalURL(req.URL); err != nil {
		return nil, "", fmt.Errorf("URL不合法: %s", err.Error())
	}

	// 用户链接数限制（max_links）
	if s.userRepo != nil && s.linkRepo != nil {
		u, err := s.userRepo.GetUserByID(ctx, userID)
		if err == nil && u.MaxLinks != -1 {
			cnt, err := s.linkRepo.CountLinksByUser(ctx, userID)
			if err == nil && int64(u.MaxLinks) <= cnt {
				return nil, "", fmt.Errorf("已达到最大链接数限制（%d条），请联系管理员", u.MaxLinks)
			}
		}
	}

	// 选择域名
	var domainID int64
	var domain *models.Domain
	if req.DomainID > 0 {
		d, err := s.domainRepo.GetDomainByID(ctx, req.DomainID)
		if err != nil {
			return nil, "", fmt.Errorf("域名不存在或无权限")
		}
		if !d.IsActive {
			return nil, "", fmt.Errorf("域名已停用")
		}
		if d.UserID != userID && d.UserID != 0 {
			return nil, "", fmt.Errorf("域名不存在或无权限")
		}
		domain = d
		domainID = d.ID
	} else {
		d, err := s.domainRepo.GetDefaultDomain(ctx, userID)
		if err == nil && d != nil {
			domain = d
			domainID = d.ID
		}
	}

	hash := s.GenerateHash(req.URL)

	// 幂等：同一 user + domain + hash 返回已存在链接
	if existing, err := s.linkRepo.GetLinkByHashUserDomain(ctx, hash, userID, domainID); err == nil && existing != nil {
		shortURL := s.BuildShortURL(domain, existing.Code)
		return existing, shortURL, nil
	}

	now := time.Now()

	// 自定义 code：先查再写（最终仍以唯一约束为准）
	if strings.TrimSpace(req.Code) != "" {
		code := strings.TrimSpace(req.Code)
		exists, err := s.linkRepo.CheckCodeExistsInDomain(ctx, code, domainID)
		if err != nil {
			return nil, "", fmt.Errorf("检查代码失败: %w", err)
		}
		if exists {
			return nil, "", fmt.Errorf("代码 %s 已存在", code)
		}
		shortURL := s.BuildShortURL(domain, code)
		qr, _ := utils.GenerateQRCode(shortURL, 256)
		link := &models.Link{
			UserID:      userID,
			DomainID:    domainID,
			Code:        code,
			OriginalURL: req.URL,
			Title:       req.Title,
			Hash:        hash,
			QRCode:      qr,
			ClickCount:  0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := s.linkRepo.CreateLink(ctx, link); err != nil {
			if repo.IsUniqueViolation(err) {
				return nil, "", fmt.Errorf("代码 %s 已存在", code)
			}
			return nil, "", fmt.Errorf("创建链接失败: %w", err)
		}
		return link, shortURL, nil
	}

	// 随机 code：并发安全重试，必要时递增长度
	length, err := s.GetAvailableCodeLength(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("获取代码长度失败: %w", err)
	}
	_, max, _ := s.getMinMaxCodeLength(ctx)

	attempts := 0
	for {
		attempts++
		code := s.GenerateRandomCode(length)
		shortURL := s.BuildShortURL(domain, code)
		qr, _ := utils.GenerateQRCode(shortURL, 256)
		link := &models.Link{
			UserID:      userID,
			DomainID:    domainID,
			Code:        code,
			OriginalURL: req.URL,
			Title:       req.Title,
			Hash:        hash,
			QRCode:      qr,
			ClickCount:  0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := s.linkRepo.CreateLink(ctx, link)
		if err == nil {
			return link, shortURL, nil
		}
		if repo.IsUniqueViolation(err) {
			// 冲突：重试，必要时升长度
			if attempts%10 == 0 {
				length++
				if length > max {
					return nil, "", fmt.Errorf("无法生成唯一代码，已达到最大长度")
				}
			}
			if attempts < 50 {
				continue
			}
			return nil, "", fmt.Errorf("生成短链冲突过多，请稍后重试")
		}
		return nil, "", fmt.Errorf("创建链接失败: %w", err)
	}
}


