import request from '@/utils/request'

// ========== 网关服务器 ==========
export function listGatewayServers() {
  return request({ url: '/gateway/servers', method: 'get' })
}
export function getGatewayServer(id) {
  return request({ url: `/gateway/servers/${id}`, method: 'get' })
}
export function createGatewayServer(data) {
  return request({ url: '/gateway/servers', method: 'post', data })
}
export function updateGatewayServer(id, data) {
  return request({ url: `/gateway/servers/${id}`, method: 'put', data })
}
export function deleteGatewayServer(id) {
  return request({ url: `/gateway/servers/${id}`, method: 'delete' })
}
export function startGatewayServer(id) {
  return request({ url: `/gateway/servers/${id}/start`, method: 'post' })
}
export function stopGatewayServer(id) {
  return request({ url: `/gateway/servers/${id}/stop`, method: 'post' })
}

// ========== 运行时链路 ==========
export function listGatewayConnections(serverId) {
  if (serverId == null) {
    return request({ url: '/gateway/connections', method: 'get' })
  }
  return request({ url: `/gateway/servers/${serverId}/connections`, method: 'get' })
}
export function getGatewayConnection(serverId, connId) {
  return request({ url: `/gateway/servers/${serverId}/connections/${connId}`, method: 'get' })
}
export function setGatewayPoll(serverId, connId, poll) {
  return request({ url: `/gateway/servers/${serverId}/connections/${connId}/poll`, method: 'put', data: poll })
}

// ========== 链路设备（Device.LinkID = link_id） ==========
export function listLinkDevices(linkId, params) {
  return request({ url: `/gateway/links/${encodeURIComponent(linkId)}/devices`, method: 'get', params })
}
export function createLinkDevice(linkId, data) {
  return request({ url: `/gateway/links/${encodeURIComponent(linkId)}/devices`, method: 'post', data })
}
export function updateLinkDevice(linkId, deviceId, data) {
  return request({
    url: `/gateway/links/${encodeURIComponent(linkId)}/devices/${deviceId}`,
    method: 'put',
    data
  })
}
export function deleteLinkDevice(linkId, deviceId) {
  return request({
    url: `/gateway/links/${encodeURIComponent(linkId)}/devices/${deviceId}`,
    method: 'delete'
  })
}
