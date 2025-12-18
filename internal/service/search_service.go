/**
 * 搜索服务（重写版，Meilisearch）
 * - 提供基于 Meilisearch 的链接搜索能力
 */
package service

import (
	"context"
	"fmt"
	"short-link/internal/config"
	"short-link/models"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

// SearchService 搜索服务
type SearchService struct {
	cfg    *config.Config
	client *meilisearch.Client
	index  *meilisearch.Index
}

// NewSearchService 创建搜索服务实例（v2）
func NewSearchService(cfg *config.Config) (*SearchService, error) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   cfg.MeiliHost,
		APIKey: cfg.MeiliKey,
	})

	if _, err := client.Health(); err != nil {
		return nil, fmt.Errorf("Meilisearch连接失败: %w", err)
	}

	index := client.Index("links")
	return &SearchService{
		cfg:    cfg,
		client: client,
		index:  index,
	}, nil
}

// SearchLinks 搜索链接（仅按 code/title/original_url）
func (s *SearchService) SearchLinks(ctx context.Context, query string, page, limit int) (*models.PaginatedLinksResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	req := &meilisearch.SearchRequest{
		Query:  query,
		Limit:  int64(limit),
		Offset: int64((page - 1) * limit),
		Sort:   []string{"created_at:desc"},
	}

	result, err := s.index.Search(query, req)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	var links []models.LinkResponse
	for _, hit := range result.Hits {
		doc, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		link := models.LinkResponse{
			ID:          int64(doc["id"].(float64)),
			Code:        doc["code"].(string),
			OriginalURL: doc["original_url"].(string),
			Title:       getString(doc, "title"),
		}

		// 短链URL：使用 BaseURL + code（与旧版保持一致）
		link.ShortURL = fmt.Sprintf("%s/%s", s.cfg.BaseURL, link.Code)

		if createdAt, ok := doc["created_at"].(float64); ok {
			link.CreatedAt = time.Unix(int64(createdAt), 0).Format(time.RFC3339)
		}

		links = append(links, link)
	}

	totalPages := int(result.EstimatedTotalHits) / limit
	if int(result.EstimatedTotalHits)%limit > 0 {
		totalPages++
	}

	return &models.PaginatedLinksResponse{
		Links:      links,
		Total:      result.EstimatedTotalHits,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}


