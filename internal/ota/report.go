package ota

import (
	"time"

	"github.com/god-jason/iot-master/internal/types"
	"github.com/god-jason/iot-master/pkg/db"
	"github.com/god-jason/iot-master/pkg/logger"

	"xorm.io/builder"
)

// HandleReport 处理设备 OTA 进度上报（由内置 broker / 外部 MQTT 驱动调用）。
// stage: downloading / verifying / installing / rebooting / success / failed / rolled_back / cancelled
func HandleReport(deviceID, taskID, stage string, progress int, errMsg, version string) {
	if taskID == "" || deviceID == "" {
		return
	}

	var row types.OTATaskDevice
	found, err := db.FindOne("ota_task_device",
		builder.And(builder.Eq{"task_id": taskID}, builder.Eq{"device_id": deviceID}), &row)
	if err != nil {
		logger.Warn("OTA 上报查询明细失败", "task", taskID, "device", deviceID, "error", err)
		return
	}
	if !found {
		logger.Warn("OTA 上报未匹配任务设备", "task", taskID, "device", deviceID)
		return
	}

	now := time.Now()
	fields := map[string]interface{}{"updated": now}
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	if progress > 0 || stage == types.OTADeviceStatusSuccess {
		fields["progress"] = progress
	}
	if errMsg != "" {
		fields["error"] = truncate(errMsg, 500)
	}

	switch stage {
	case types.OTADeviceStatusDownloading,
		types.OTADeviceStatusVerifying,
		types.OTADeviceStatusInstalling,
		types.OTADeviceStatusRebooting,
		types.OTADeviceStatusSuccess,
		types.OTADeviceStatusFailed,
		types.OTADeviceStatusRolledBack,
		types.OTADeviceStatusCancelled:
		fields["status"] = stage
	default:
		// 未知阶段仅更新进度与时间，不改变状态
	}

	if terminalDeviceStatus(stage) {
		fields["finished_at"] = now
	}

	updateTaskDevice(row.ID, fields)

	if stage == types.OTADeviceStatusSuccess {
		newVersion := version
		if newVersion == "" {
			newVersion = row.ToVersion
		}
		if newVersion != "" {
			if _, err := db.UpdateOne("device", builder.Eq{"id": deviceID}, map[string]interface{}{
				"firmware_version":    newVersion,
				"firmware_updated_at": now,
			}); err != nil {
				logger.Warn("OTA 升级成功回写设备版本失败", "device", deviceID, "error", err)
			}
		}
	}

	refreshTaskCounts(taskID)
	logger.Info("OTA 进度上报", "task", taskID, "device", deviceID, "stage", stage, "progress", progress)
}
