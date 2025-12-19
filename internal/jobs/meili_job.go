/**
 * Meilisearch 写入任务队列与 Worker（异步写入 + 失败重试）
 * 实现 redo.md 2.6：Meilisearch 写入失败补偿/重试/后台任务
 */
package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"short-link/internal/config"
	"short-link/models"
	"short-link/utils"

	"github.com/meilisearch/meilisearch-go"
)

// MeiliTask Meilisearch 写入任务
type MeiliTask struct {
	Action string      `json:"action"` // "index", "delete"
	Link   *models.Link `json:"link,omitempty"`
	LinkID int64       `json:"link_id,omitempty"`
}

// MeiliWorker Meilisearch 写入 Worker
type MeiliWorker struct {
	taskChan    chan *MeiliTask
	client      *meilisearch.Client
	index       *meilisearch.Index
	maxRetries  int
	retryDelay  time.Duration
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewMeiliWorker 创建 Meilisearch Worker
func NewMeiliWorker(cfg *config.Config, maxRetries int, retryDelay time.Duration) (*MeiliWorker, error) {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   cfg.MeiliHost,
		APIKey: cfg.MeiliKey,
	})

	// 测试连接
	if _, err := client.Health(); err != nil {
		return nil, fmt.Errorf("Meilisearch连接失败: %w", err)
	}

	index := client.Index("links")
	ctx, cancel := context.WithCancel(context.Background())

	return &MeiliWorker{
		taskChan:   make(chan *MeiliTask, 1000), // 缓冲1000个任务
		client:     client,
		index:      index,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Submit 提交 Meilisearch 写入任务（非阻塞）
func (w *MeiliWorker) Submit(action string, link *models.Link, linkID int64) {
	task := &MeiliTask{
		Action: action,
		Link:   link,
		LinkID: linkID,
	}

	select {
	case w.taskChan <- task:
		// 成功提交
	default:
		// 队列满，静默丢弃（避免阻塞主流程）
		utils.LogWarn("Meilisearch任务队列已满，丢弃任务: action=%s, link_id=%d", action, linkID)
	}
}

// Start 启动 Worker（后台 goroutine）
func (w *MeiliWorker) Start() {
	w.wg.Add(1)
	go w.run()
	utils.LogInfo("Meilisearch Worker 已启动（最大重试次数=%d，重试间隔=%v）", w.maxRetries, w.retryDelay)
}

// run Worker 主循环
func (w *MeiliWorker) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return

		case task := <-w.taskChan:
			w.processTask(task)
		}
	}
}

// processTask 处理单个任务（带重试）
func (w *MeiliWorker) processTask(task *MeiliTask) {
	var err error
	for attempt := 0; attempt < w.maxRetries; attempt++ {
		if attempt > 0 {
			// 重试前等待
			select {
			case <-w.ctx.Done():
				return
			case <-time.After(w.retryDelay * time.Duration(attempt)):
			}
		}

		switch task.Action {
		case "index":
			err = w.indexLink(task.Link)
		case "delete":
			err = w.deleteLink(task.LinkID)
		default:
			utils.LogError("未知的 Meilisearch 任务类型: %s", task.Action)
			return
		}

		if err == nil {
			// 成功
			return
		}

		utils.LogWarn("Meilisearch写入失败（尝试 %d/%d）: action=%s, link_id=%d, error=%v",
			attempt+1, w.maxRetries, task.Action, task.LinkID, err)
	}

	// 所有重试都失败，记录错误（后续可扩展为死信队列）
	utils.LogError("Meilisearch写入最终失败: action=%s, link_id=%d, error=%v", task.Action, task.LinkID, err)
}

// indexLink 索引链接到 Meilisearch
func (w *MeiliWorker) indexLink(link *models.Link) error {
	if link == nil {
		return fmt.Errorf("link 不能为空")
	}

	doc := map[string]interface{}{
		"id":           link.ID,
		"code":         link.Code,
		"original_url": link.OriginalURL,
		"title":        link.Title,
		"user_id":      link.UserID,
		"domain_id":    link.DomainID,
		"created_at":   link.CreatedAt.Unix(),
	}

	_, err := w.index.AddDocuments([]map[string]interface{}{doc}, "id")
	return err
}

// deleteLink 从 Meilisearch 删除链接
func (w *MeiliWorker) deleteLink(linkID int64) error {
	_, err := w.index.DeleteDocument(fmt.Sprintf("%d", linkID))
	return err
}

// Stop 停止 Worker
func (w *MeiliWorker) Stop() {
	w.cancel()
	w.wg.Wait()
	utils.LogInfo("Meilisearch Worker 已停止")
}


