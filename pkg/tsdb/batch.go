package tsdb

import (
	"context"
	"sync"
	"time"

	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/tsdbapi"
)

const (
	batchSize     = 200
	flushInterval = 500 * time.Millisecond
	queueSize     = 8192
)

var (
	batchQueue = make(chan tsdbapi.Point, queueSize)
	batchOnce  sync.Once
)

// Write 异步批量写入单条数据点。
// 设备上报主链路只入队、不等待落库；后台协程按 batchSize 或 flushInterval 触发出队，
// 调用 Store.WriteBatch 一次性写入，避免每条上报都触发一次 IO。
// 队列满时丢弃并打 Warn 日志，保证不会阻塞设备上报。
func Write(pt tsdbapi.Point) {
	if DefaultStore == nil {
		return
	}
	batchOnce.Do(startBatchWriter)

	select {
	case batchQueue <- pt:
	default:
		logger.Warn("tsdb 写入队列已满，丢弃数据点", "measurement", pt.Measurement)
	}
}

func startBatchWriter() {
	go func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		batch := make([]tsdbapi.Point, 0, batchSize)
		flush := func() {
			if len(batch) == 0 {
				return
			}
			store := DefaultStore
			if store != nil {
				if err := store.WriteBatch(context.Background(), batch); err != nil {
					logger.Warn("tsdb 批量写入失败", "count", len(batch), "error", err)
				}
			}
			batch = batch[:0]
		}

		for {
			select {
			case pt := <-batchQueue:
				batch = append(batch, pt)
				if len(batch) >= batchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
}
