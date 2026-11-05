package ota

import (
	"sort"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"xorm.io/builder"
)

// worker OTA 调度器：定时扫描任务、推进批次、下发指令、处理超时与重试。
type worker struct {
	stop chan struct{}
	sem  chan struct{} // 全局下发并发令牌
}

var current *worker

// Start boot 启动入口
func Start() error {
	if !Enabled() {
		logger.Info("OTA 模块已禁用")
		return nil
	}
	w := &worker{
		stop: make(chan struct{}),
		sem:  make(chan struct{}, Concurrency()),
	}
	current = w
	go w.loop()
	logger.Info("OTA 升级调度器已启动", "scan_interval", ScanInterval().String(), "concurrency", Concurrency())
	return nil
}

// Stop boot 关闭入口
func Stop() error {
	if current != nil {
		close(current.stop)
		current = nil
	}
	return nil
}

func (w *worker) loop() {
	ticker := time.NewTicker(ScanInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.scan()
		case <-w.stop:
			return
		}
	}
}

func (w *worker) scan() {
	var tasks []types.OTATask
	cond := builder.Or(
		builder.Eq{"status": types.OTATaskStatusScheduled},
		builder.Eq{"status": types.OTATaskStatusRunning},
	)
	if err := db.Find("ota_task", cond, &tasks, 0, 0); err != nil {
		logger.Warn("OTA 扫描任务失败", "error", err)
		return
	}
	for i := range tasks {
		w.processTask(&tasks[i])
	}
}

// processTask 推进单个任务：到点启动、激活批次、下发设备、超时处理、批次完成与任务收尾。
func (w *worker) processTask(task *types.OTATask) {
	now := time.Now()

	// pending scheduled → running
	if task.Status == types.OTATaskStatusScheduled {
		if !task.ScheduledAt.IsZero() && task.ScheduledAt.After(now) {
			return
		}
		updateTask(task.ID, map[string]interface{}{
			"status":     types.OTATaskStatusRunning,
			"started_at": now,
		})
		task.Status = types.OTATaskStatusRunning
	}
	if task.Status != types.OTATaskStatusRunning {
		return
	}

	// 固件必须仍处于已发布状态
	var fw types.Firmware
	found, err := db.FindOne("ota_firmware", builder.Eq{"id": task.FirmwareID}, &fw)
	if err != nil || !found || fw.ID == "" {
		logger.Warn("OTA 任务固件不存在，暂停任务", "task", task.ID, "firmware", task.FirmwareID)
		pauseTask(task, "auto")
		return
	}
	if fw.Status != types.FirmwareStatusReleased {
		logger.Warn("OTA 任务固件已停用，暂停任务", "task", task.ID, "firmware", task.FirmwareID)
		pauseTask(task, "auto")
		return
	}

	batches := loadBatches(task.ID)
	if len(batches) == 0 {
		finalizeTask(task)
		return
	}

	active := findRunningBatch(batches)
	if active == nil {
		next, prev := findNextBatch(batches)
		if next == nil {
			finalizeTask(task)
			return
		}
		// 非首批：必须等上一批完成、通过门禁并满足等待时间
		if prev != nil {
			if prev.Status != types.OTABatchStatusCompleted {
				return
			}
			if !gatePassed(task, prev) {
				pauseTask(task, "auto")
				return
			}
			if wait := time.Duration(next.WaitMinutes) * time.Minute; now.Before(prev.FinishedAt.Add(wait)) {
				return
			}
		}
		updateBatchStatus(task.ID, next.Seq, types.OTABatchStatusRunning, now, time.Time{})
		next.Status = types.OTABatchStatusRunning
		next.StartedAt = now
		active = next
	}

	w.processBatch(task, &fw, active)
}

// processBatch 处理一个已激活批次：超时检查、完成后收尾、否则下发在线设备。
func (w *worker) processBatch(task *types.OTATask, fw *types.Firmware, batch *types.OTATaskBatch) {
	now := time.Now()
	var rows []types.OTATaskDevice
	if err := db.Find("ota_task_device", builder.Eq{"task_id": task.ID, "batch_seq": batch.Seq}, &rows, 0, 0); err != nil {
		logger.Warn("OTA 读取批次设备失败", "task", task.ID, "error", err)
		return
	}
	if len(rows) == 0 {
		updateBatchStatus(task.ID, batch.Seq, types.OTABatchStatusCompleted, batch.StartedAt, now)
		return
	}

	timeout := ReportTimeout()
	pending := make([]types.OTATaskDevice, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		if terminalDeviceStatus(row.Status) {
			continue
		}
		// 只对已下发且进入升级流程的设备做无进度超时判断；pending 等待下发不计时
		if row.Status != types.OTADeviceStatusPending {
			last := lastActivity(row)
			if !last.IsZero() && now.Sub(last) > timeout {
				w.handleTimeout(task, row)
				continue
			}
		}
		if row.Status == types.OTADeviceStatusPending {
			pending = append(pending, *row)
		}
	}

	// 本批全部终态 → 标记批次完成，下一轮扫描推进
	total, _, _, terminal, err := countTaskDevices(task.ID, batch.Seq)
	if err == nil && terminal >= total && total > 0 {
		updateBatchStatus(task.ID, batch.Seq, types.OTABatchStatusCompleted, batch.StartedAt, time.Now())
		refreshTaskCounts(task.ID)
		return
	}

	if task.Status != types.OTATaskStatusRunning {
		return
	}
	w.dispatchPending(task, fw, pending)
}

// dispatchPending 对在线设备逐个尝试下发；用 DB 状态的条件更新做抢占，避免重复下发。
func (w *worker) dispatchPending(task *types.OTATask, fw *types.Firmware, pending []types.OTATaskDevice) {
	for i := range pending {
		row := &pending[i]
		dev, ok := iot.GetDevice(row.DeviceID)
		if !ok || !dev.IsOnline() {
			continue
		}
		// 同一设备不允许同时存在两个未终态任务
		busy, err := ActiveDeviceIDs([]string{row.DeviceID}, task.ID)
		if err != nil {
			continue
		}
		if busy[row.DeviceID] {
			continue
		}

		now := time.Now()
		affected, err := db.UpdateOne("ota_task_device",
			builder.And(
				builder.Eq{"id": row.ID},
				builder.Eq{"status": types.OTADeviceStatusPending},
			),
			map[string]interface{}{
				"status":        types.OTADeviceStatusDispatched,
				"dispatched_at": now,
				"updated":       now,
			})
		if err != nil || affected == 0 {
			continue
		}

		// 记录升级前版本（供日志展示）
		var devDB types.Device
		if found, _ := db.FindOne("device", builder.Eq{"id": row.DeviceID}, &devDB); found {
			updateTaskDevice(row.ID, map[string]interface{}{"from_version": devDB.FirmwareVersion})
		}

		go w.dispatch(task, fw, row.ID, row.DeviceID)
	}
}

// dispatch 实际下发 ota_update 服务调用并等待 ACK；失败按重试次数重排或置为失败。
func (w *worker) dispatch(task *types.OTATask, fw *types.Firmware, rowID int64, deviceID string) {
	w.sem <- struct{}{}
	defer func() { <-w.sem }()

	dev, ok := iot.GetDevice(deviceID)
	if !ok || !dev.IsOnline() {
		updateTaskDevice(rowID, map[string]interface{}{
			"status":  types.OTADeviceStatusPending,
			"updated": time.Now(),
		})
		return
	}

	url := DownloadURL(fw)
	if url == "" || fw.SHA256 == "" {
		w.failOrRetry(task.ID, rowID, "固件下载地址或 SHA256 未配置")
		return
	}

	params := map[string]interface{}{
		"task_id":       task.ID,
		"url":           url,
		"size":          fw.FileSize,
		"sha256":        fw.SHA256,
		"version":       fw.Version,
		"min_version":   fw.MinVersion,
		"signature":     fw.Signature,
		"auto_rollback": task.AutoRollback,
	}
	if _, err := dev.Control("ota_update", params); err != nil {
		w.failOrRetry(task.ID, rowID, err.Error())
		return
	}

	now := time.Now()
	updateTaskDevice(rowID, map[string]interface{}{
		"status":        types.OTADeviceStatusDispatched,
		"dispatched_at": now,
		"updated":       now,
		"error":         "",
	})
	logger.Info("OTA 升级指令已下发", "task", task.ID, "device", deviceID, "version", fw.Version)
}

// handleTimeout 无进度超时：未超过重试上限则回到 pending，否则置为 timeout。
func (w *worker) handleTimeout(task *types.OTATask, row *types.OTATaskDevice) {
	retry := row.RetryCount + 1
	if retry <= MaxRetry() {
		updateTaskDevice(row.ID, map[string]interface{}{
			"status":      types.OTADeviceStatusPending,
			"retry_count": retry,
			"error":       "升级超时，等待重试",
			"updated":     time.Now(),
		})
		logger.Warn("OTA 设备升级超时，安排重试", "task", task.ID, "device", row.DeviceID, "retry", retry)
		refreshTaskCounts(task.ID)
		return
	}
	updateTaskDevice(row.ID, map[string]interface{}{
		"status":      types.OTADeviceStatusTimeout,
		"retry_count": retry,
		"error":       "升级超时",
		"finished_at": time.Now(),
		"updated":     time.Now(),
	})
	logger.Warn("OTA 设备升级超时", "task", task.ID, "device", row.DeviceID)
	refreshTaskCounts(task.ID)
}

// failOrRetry 下发失败：未超过重试上限则回到 pending，否则置为 failed。
func (w *worker) failOrRetry(taskID string, rowID int64, msg string) {
	var row types.OTATaskDevice
	found, err := db.FindOne("ota_task_device", builder.Eq{"id": rowID}, &row)
	if err != nil || !found || terminalDeviceStatus(row.Status) {
		return
	}
	retry := row.RetryCount + 1
	if retry <= MaxRetry() {
		updateTaskDevice(rowID, map[string]interface{}{
			"status":      types.OTADeviceStatusPending,
			"retry_count": retry,
			"error":       truncate(msg, 500),
			"updated":     time.Now(),
		})
		refreshTaskCounts(taskID)
		return
	}
	updateTaskDevice(rowID, map[string]interface{}{
		"status":      types.OTADeviceStatusFailed,
		"retry_count": retry,
		"error":       truncate(msg, 500),
		"finished_at": time.Now(),
		"updated":     time.Now(),
	})
	logger.Warn("OTA 设备下发失败", "task", taskID, "device", row.DeviceID, "error", msg)
	refreshTaskCounts(taskID)
}

// gatePassed 判断上一批是否达到灰度门禁；失败率超限或成功率不足时返回 false。
func gatePassed(task *types.OTATask, prev *types.OTATaskBatch) bool {
	st := ParseStrategy(task.Strategy)
	_, success, failed, terminal, err := countTaskDevices(task.ID, prev.Seq)
	if err != nil || terminal == 0 {
		return true
	}
	failRate := float64(failed) / float64(terminal)
	if failRate > st.FailureThreshold {
		logger.Warn("OTA 批次失败率超限，暂停任务", "task", task.ID, "batch", prev.Seq, "fail_rate", failRate)
		return false
	}
	successRate := float64(success) / float64(terminal)
	if successRate < st.SuccessThreshold {
		logger.Warn("OTA 批次成功率不足，暂停任务", "task", task.ID, "batch", prev.Seq, "success_rate", successRate)
		return false
	}
	return true
}

// pauseTask 自动暂停任务
func pauseTask(task *types.OTATask, reason string) {
	updateTask(task.ID, map[string]interface{}{
		"status":       types.OTATaskStatusPaused,
		"pause_reason": reason,
	})
}

// finalizeTask 所有批次结束后汇总任务状态
func finalizeTask(task *types.OTATask) {
	total, success, failed, _, err := countTaskDevices(task.ID, 0)
	if err != nil {
		return
	}
	status := types.OTATaskStatusFailed
	switch {
	case total == 0:
		status = types.OTATaskStatusFailed
	case success == total:
		status = types.OTATaskStatusCompleted
	case success > 0:
		status = types.OTATaskStatusPartial
	}
	if task.Status == types.OTATaskStatusCancelled {
		status = types.OTATaskStatusCancelled
	}
	updateTask(task.ID, map[string]interface{}{
		"status":      status,
		"success":     success,
		"failed":      failed,
		"finished_at": time.Now(),
	})
	logger.Info("OTA 任务结束", "task", task.ID, "status", status, "total", total, "success", success, "failed", failed)
}

// refreshTaskCounts 实时刷新任务成功/失败数（进度上报时调用，便于后台即时看到统计）
func refreshTaskCounts(taskID string) {
	_, success, failed, _, err := countTaskDevices(taskID, 0)
	if err != nil {
		return
	}
	updateTask(taskID, map[string]interface{}{
		"success": success,
		"failed":  failed,
	})
}

// loadBatches 读取任务批次并按序号排序
func loadBatches(taskID string) []types.OTATaskBatch {
	var batches []types.OTATaskBatch
	if err := db.Find("ota_task_batch", builder.Eq{"task_id": taskID}, &batches, 0, 0); err != nil {
		return nil
	}
	sort.SliceStable(batches, func(i, j int) bool { return batches[i].Seq < batches[j].Seq })
	return batches
}

// findRunningBatch 返回当前正在执行的批次
func findRunningBatch(batches []types.OTATaskBatch) *types.OTATaskBatch {
	for i := range batches {
		if batches[i].Status == types.OTABatchStatusRunning {
			return &batches[i]
		}
	}
	// 若批次被异常中断（状态 pending 但已有设备下发），保持 pending 处理
	return nil
}

// findNextBatch 返回第一个待执行批次及其上一批
func findNextBatch(batches []types.OTATaskBatch) (next *types.OTATaskBatch, prev *types.OTATaskBatch) {
	for i := range batches {
		if batches[i].Status == types.OTABatchStatusPending {
			next = &batches[i]
			if i > 0 {
				prev = &batches[i-1]
			}
			return
		}
		if batches[i].Status == types.OTABatchStatusRunning {
			return nil, nil
		}
	}
	return nil, nil
}

func updateTask(taskID string, fields map[string]interface{}) {
	if _, err := db.UpdateOne("ota_task", builder.Eq{"id": taskID}, fields); err != nil {
		logger.Warn("更新 OTA 任务失败", "task", taskID, "error", err)
	}
}

func updateBatchStatus(taskID string, seq int, status string, startedAt, finishedAt time.Time) {
	fields := map[string]interface{}{"status": status}
	if !startedAt.IsZero() {
		fields["started_at"] = startedAt
	}
	if !finishedAt.IsZero() {
		fields["finished_at"] = finishedAt
	}
	if _, err := db.UpdateOne("ota_task_batch",
		builder.And(builder.Eq{"task_id": taskID}, builder.Eq{"seq": seq}), fields); err != nil {
		logger.Warn("更新 OTA 批次失败", "task", taskID, "seq", seq, "error", err)
	}
}

func updateTaskDevice(rowID int64, fields map[string]interface{}) {
	if _, err := db.UpdateOne("ota_task_device", builder.Eq{"id": rowID}, fields); err != nil {
		logger.Warn("更新 OTA 设备明细失败", "id", rowID, "error", err)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
