import request from '@/utils/request'

// 获取所有协议列表（代码定义，非数据库）
export function listProtocols() {
  return request({
    url: '/protocol',
    method: 'get'
  })
}

// 获取指定协议详情（按协议名，如 "tcp"、"rtu"、"cjt188"）
export function getProtocol(name) {
  return request({
    url: `/protocol/${name}`,
    method: 'get'
  })
}

// searchProtocols 兼容旧调用 —— 等价于 listProtocols（协议列表不再分页/搜索）
export function searchProtocols() {
  return listProtocols()
}
