import request from '@/utils/request'

export function getDeviceList(data) {
  return request({
    url: '/device/search',
    method: 'post',
    data
  })
}

export function getDevice(id) {
  return request({
    url: `/device/${id}`,
    method: 'get'
  })
}

export function createDevice(data) {
  return request({
    url: '/device',
    method: 'post',
    data
  })
}

export function updateDevice(id, data) {
  return request({
    url: `/device/${id}`,
    method: 'put',
    data
  })
}

export function deleteDevice(id) {
  return request({
    url: `/device/${id}`,
    method: 'delete'
  })
}

export function searchDevices(data) {
  return request({
    url: '/device/search',
    method: 'post',
    data
  })
}

export function getDeviceHistory(id, params) {
  return request({
    url: `/device/${id}/history`,
    method: 'get',
    params
  })
}

export function getDeviceProperty(id) {
  return request({
    url: `/device/${id}/property`,
    method: 'get'
  })
}

export function setDeviceConfig(id, data) {
  return request({
    url: `/device/${id}/setting/set`,
    method: 'post',
    data
  })
}

export function executeDeviceControl(id, service, params) {
  return request({
    url: `/device/${id}/control`,
    method: 'post',
    data: { service, params }
  })
}

export function searchDeviceLogs(data) {
  return request({
    url: '/device-log/search',
    method: 'post',
    data
  })
}

export function getDeviceEventList(data) {
  return request({
    url: '/device-event/search',
    method: 'post',
    data
  })
}

export function deleteDeviceEvent(id) {
  return request({
    url: `/device-event/${id}`,
    method: 'delete'
  })
}

export function handleDeviceEvent(id) {
  return request({
    url: `/device-event/${id}/handle`,
    method: 'put'
  })
}

export function unbindDevice(id, bindingId) {
  return request({
    url: `/device/${id}/unbind/${bindingId}`,
    method: 'delete'
  })
}

// ========== 设备共享/绑定 ==========

// 分享设备给其他用户
export function shareDevice(id, data) {
  return request({
    url: `/device/${id}/share`,
    method: 'post',
    data
  })
}

// 查询设备绑定关系列表
export function getDeviceBindings(id) {
  return request({
    url: `/device/${id}/bindings`,
    method: 'get'
  })
}
