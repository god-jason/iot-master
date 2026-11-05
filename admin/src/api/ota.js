import request from '@/utils/request'

// ========== 固件管理 ==========

export function searchFirmware(data) {
  return request({
    url: '/firmware/search',
    method: 'post',
    data
  })
}

export function getFirmware(id) {
  return request({
    url: `/firmware/${id}`,
    method: 'get'
  })
}

export function updateFirmware(id, data) {
  return request({
    url: `/firmware/${id}`,
    method: 'put',
    data
  })
}

export function deleteFirmware(id) {
  return request({
    url: `/firmware/${id}`,
    method: 'delete'
  })
}

export function releaseFirmware(id) {
  return request({
    url: `/firmware/${id}/release`,
    method: 'put'
  })
}

export function disableFirmware(id) {
  return request({
    url: `/firmware/${id}/disable`,
    method: 'put'
  })
}

// 固件黑白名单
export function listFirmwareRules(id) {
  return request({
    url: `/firmware/${id}/rules`,
    method: 'get'
  })
}

// 添加名单规则：{ list_type, target_type, target_ids }
export function addFirmwareRules(id, data) {
  return request({
    url: `/firmware/${id}/rules`,
    method: 'post',
    data
  })
}

export function deleteFirmwareRule(id, ruleId) {
  return request({
    url: `/firmware/${id}/rules/${ruleId}`,
    method: 'delete'
  })
}

// 上传固件（FormData: file + name/version/model_id/channel/min_version/signature/notes）
export function uploadFirmware(formData, onUploadProgress) {
  return request({
    url: '/upload/firmware',
    method: 'post',
    data: formData,
    timeout: 120000,
    onUploadProgress
  })
}

// ========== 升级任务 ==========

export function previewOTATask(data) {
  return request({
    url: '/ota/task/preview',
    method: 'post',
    data
  })
}

export function createOTATask(data) {
  return request({
    url: '/ota/task',
    method: 'post',
    data
  })
}

export function searchOTATasks(data) {
  return request({
    url: '/ota/task/search',
    method: 'post',
    data
  })
}

export function getOTATask(id) {
  return request({
    url: `/ota/task/${id}`,
    method: 'get'
  })
}

export function getOTATaskDevices(id, params) {
  return request({
    url: `/ota/task/${id}/devices`,
    method: 'get',
    params
  })
}

export function pauseOTATask(id) {
  return request({
    url: `/ota/task/${id}/pause`,
    method: 'put'
  })
}

export function resumeOTATask(id) {
  return request({
    url: `/ota/task/${id}/resume`,
    method: 'put'
  })
}

export function cancelOTATask(id) {
  return request({
    url: `/ota/task/${id}/cancel`,
    method: 'put'
  })
}

export function retryFailedOTATask(id) {
  return request({
    url: `/ota/task/${id}/retry-failed`,
    method: 'put'
  })
}

// 单设备升级
export function upgradeDevice(deviceId, data) {
  return request({
    url: `/device/${deviceId}/ota`,
    method: 'post',
    data
  })
}
