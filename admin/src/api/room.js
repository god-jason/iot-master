import request from '@/utils/request'

// ========== 房间管理 ==========

// 搜索房间列表
export function searchRooms(data) {
  return request({ url: '/room/search', method: 'post', data })
}

// 创建房间
export function createRoom(data) {
  return request({ url: '/room', method: 'post', data })
}

// 更新房间
export function updateRoom(id, data) {
  return request({ url: `/room/${id}`, method: 'put', data })
}

// 删除房间
export function deleteRoom(id) {
  return request({ url: `/room/${id}`, method: 'delete' })
}

// ========== 设备日志 ==========

// 搜索设备日志
export function searchDeviceLogs(data) {
  return request({ url: '/device-log/search', method: 'post', data })
}

// 获取单条设备日志详情
export function getDeviceLog(id) {
  return request({ url: `/device-log/${id}`, method: 'get' })
}
