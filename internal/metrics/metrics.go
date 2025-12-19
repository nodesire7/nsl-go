/**
 * Prometheus Metrics
 * 实现 redo.md 6.3：Metrics（Prometheus）指标收集和暴露
 */
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP 请求总数（按方法、路径、状态码）
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP 请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP 请求延迟（按方法、路径）
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP 请求延迟（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// 短链接创建总数
	LinksCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "links_created_total",
			Help: "创建的短链接总数",
		},
	)

	// 短链接重定向总数
	LinksRedirectedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "links_redirected_total",
			Help: "短链接重定向总数",
		},
	)

	// 短链接删除总数
	LinksDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "links_deleted_total",
			Help: "删除的短链接总数",
		},
	)

	// 数据库查询延迟（按操作类型）
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "数据库查询延迟（秒）",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"operation"},
	)

	// Redis 缓存命中率
	RedisCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Redis 缓存命中总数",
		},
	)

	RedisCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Redis 缓存未命中总数",
		},
	)

	// Meilisearch 写入成功/失败
	MeilisearchWritesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "meilisearch_writes_total",
			Help: "Meilisearch 写入总数",
		},
		[]string{"status"}, // "success" or "failure"
	)

	// 限流拒绝的请求数
	RateLimitRejectedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_rejected_total",
			Help: "被限流拒绝的请求总数",
		},
	)

	// 当前活跃用户数（Gauge）
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "当前活跃用户数",
		},
	)

	// 当前短链接总数（Gauge）
	TotalLinks = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_links",
			Help: "当前短链接总数",
		},
	)
)

