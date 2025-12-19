/**
 * 统计任务队列与 Worker（异步写入访问日志/点击数）
 * 实现 redo.md 5.3：跳转路径必须极快，统计写入异步化
 */
package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"short-link/internal/repo"
	"short-link/models"
	"short-link/utils"
)

// StatsTask 统计任务
type StatsTask struct {
	LinkID    int64
	IP        string
	UserAgent string
	Referer   string
	CreatedAt time.Time
}

// StatsWorker 统计写入 Worker
type StatsWorker struct {
	taskChan    chan *StatsTask
	batchSize   int
	batchWait   time.Duration
	linkRepo    *repo.LinkRepo
	accessLogRepo *repo.AccessLogRepo
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewStatsWorker 创建统计 Worker
func NewStatsWorker(linkRepo *repo.LinkRepo, accessLogRepo *repo.AccessLogRepo, batchSize int, batchWait time.Duration) *StatsWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &StatsWorker{
		taskChan:     make(chan *StatsTask, 1000), // 缓冲1000个任务
		batchSize:    batchSize,
		batchWait:    batchWait,
		linkRepo:     linkRepo,
		accessLogRepo: accessLogRepo,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Submit 提交统计任务（非阻塞）
func (w *StatsWorker) Submit(linkID int64, ip, userAgent, referer string) {
	task := &StatsTask{
		LinkID:    linkID,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
		CreatedAt: time.Now(),
	}

	select {
	case w.taskChan <- task:
		// 成功提交
	default:
		// 队列满，静默丢弃（避免阻塞跳转路径）
		utils.LogWarn("统计任务队列已满，丢弃任务: link_id=%d", linkID)
	}
}

// Start 启动 Worker（后台 goroutine）
func (w *StatsWorker) Start() {
	w.wg.Add(1)
	go w.run()
	utils.LogInfo("统计 Worker 已启动（批量大小=%d，等待间隔=%v）", w.batchSize, w.batchWait)
}

// run Worker 主循环（批量处理）
func (w *StatsWorker) run() {
	defer w.wg.Done()

	batch := make([]*StatsTask, 0, w.batchSize)
	ticker := time.NewTicker(w.batchWait)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			// 关闭时处理剩余任务
			if len(batch) > 0 {
				w.flushBatch(batch)
			}
			return

		case task := <-w.taskChan:
			batch = append(batch, task)
			if len(batch) >= w.batchSize {
				w.flushBatch(batch)
				batch = batch[:0] // 重置切片但保留容量
			}

		case <-ticker.C:
			// 定时刷新（即使未满 batchSize）
			if len(batch) > 0 {
				w.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch 批量写入统计
func (w *StatsWorker) flushBatch(batch []*StatsTask) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 按 link_id 分组，合并点击数
	clickCounts := make(map[int64]int)
	accessLogs := make([]*models.AccessLog, 0, len(batch))

	for _, task := range batch {
		clickCounts[task.LinkID]++
		accessLogs = append(accessLogs, &models.AccessLog{
			LinkID:    task.LinkID,
			IP:        task.IP,
			UserAgent: task.UserAgent,
			Referer:   task.Referer,
			CreatedAt: task.CreatedAt,
		})
	}

	// 批量写入点击数（使用事务或批量 UPDATE）
	for linkID, count := range clickCounts {
		if err := w.linkRepo.IncrementClickCount(ctx, linkID, count); err != nil {
			utils.LogError("批量写入点击数失败: link_id=%d, count=%d, error=%v", linkID, count, err)
		}
	}

	// 批量写入访问日志
	for _, log := range accessLogs {
		if err := w.accessLogRepo.CreateAccessLog(ctx, log); err != nil {
			utils.LogError("批量写入访问日志失败: link_id=%d, error=%v", log.LinkID, err)
		}
	}

	utils.LogInfo("批量写入统计完成: 点击数=%d, 访问日志=%d", len(clickCounts), len(accessLogs))
}

// Stop 停止 Worker
func (w *StatsWorker) Stop() {
	w.cancel()
	w.wg.Wait()
	utils.LogInfo("统计 Worker 已停止")
}

