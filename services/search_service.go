/**
 * 搜索服务
 * 集成Meilisearch提供全文搜索功能
 */
package services

import (
	"fmt"
	"log"
	"short-link/config"
	"short-link/models"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

// SearchService 搜索服务
type SearchService struct {
	client *meilisearch.Client
	index  *meilisearch.Index
}

// NewSearchService 创建搜索服务实例
func NewSearchService() (*SearchService, error) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   config.AppConfig.MeiliHost,
		APIKey: config.AppConfig.MeiliKey,
	})
	
	// 检查连接
	health, err := client.Health()
	if err != nil {
		return nil, fmt.Errorf("Meilisearch连接失败: %w", err)
	}
	log.Printf("Meilisearch连接成功: %s", health.Status)
	
	index := client.Index("links")
	
	// 配置索引
	searchableAttributes := []string{"code", "title", "original_url"}
	filterableAttributes := []string{"created_at", "click_count"}
	sortableAttributes := []string{"created_at", "click_count"}
	
	_, err = index.UpdateSearchableAttributes(&searchableAttributes)
	if err != nil {
		log.Printf("配置搜索属性失败: %v", err)
	}
	
	_, err = index.UpdateFilterableAttributes(&filterableAttributes)
	if err != nil {
		log.Printf("配置过滤属性失败: %v", err)
	}
	
	_, err = index.UpdateSortableAttributes(&sortableAttributes)
	if err != nil {
		log.Printf("配置排序属性失败: %v", err)
	}
	
	return &SearchService{
		client: client,
		index:  index,
	}, nil
}

// IndexLink 索引链接
func (s *SearchService) IndexLink(link *models.Link) error {
	doc := map[string]interface{}{
		"id":           link.ID,
		"code":         link.Code,
		"original_url": link.OriginalURL,
		"title":        link.Title,
		"click_count":  link.ClickCount,
		"created_at":   link.CreatedAt.Unix(),
	}
	
	_, err := s.index.AddDocuments([]map[string]interface{}{doc}, "id")
	if err != nil {
		return fmt.Errorf("索引链接失败: %w", err)
	}
	
	log.Printf("链接已索引: %s", link.Code)
	return nil
}

// UpdateLink 更新索引中的链接
func (s *SearchService) UpdateLink(link *models.Link) error {
	return s.IndexLink(link)
}

// DeleteLink 从索引中删除链接
func (s *SearchService) DeleteLink(linkID int64) error {
	_, err := s.index.DeleteDocument(fmt.Sprintf("%d", linkID))
	if err != nil {
		return fmt.Errorf("删除索引失败: %w", err)
	}
	return nil
}

// SearchLinks 搜索链接
func (s *SearchService) SearchLinks(query string, page, limit int) (*models.PaginatedLinksResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	
	searchRequest := &meilisearch.SearchRequest{
		Query:  query,
		Limit:  int64(limit),
		Offset: int64((page - 1) * limit),
		Sort:   []string{"created_at:desc"},
	}
	
	result, err := s.index.Search(query, searchRequest)
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
			Title:       getStringValue(doc, "title"),
			ClickCount:  int64(doc["click_count"].(float64)),
		}
		
		// 设置短链接URL
		link.ShortURL = fmt.Sprintf("%s/%s", config.AppConfig.BaseURL, link.Code)
		
		// 设置创建时间
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

// getStringValue 安全获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// SyncAllLinks 同步所有链接到搜索索引
func (s *SearchService) SyncAllLinks(links []models.Link) error {
	if len(links) == 0 {
		return nil
	}
	
	var docs []map[string]interface{}
	for _, link := range links {
		doc := map[string]interface{}{
			"id":           link.ID,
			"code":         link.Code,
			"original_url": link.OriginalURL,
			"title":        link.Title,
			"click_count":  link.ClickCount,
			"created_at":   link.CreatedAt.Unix(),
		}
		docs = append(docs, doc)
	}
	
	_, err := s.index.AddDocuments(docs, "id")
	if err != nil {
		return fmt.Errorf("批量索引失败: %w", err)
	}
	
	log.Printf("已同步 %d 个链接到搜索索引", len(links))
	return nil
}

