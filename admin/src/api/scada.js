import request from '@/utils/request'

// 分页搜索画面列表（组态/大屏）
export function searchScadas(data) {
  return request({
    url: '/scada/search',
    method: 'post',
    data
  })
}

// 获取画面详情
export function getScada(id) {
  return request({
    url: `/scada/${id}`,
    method: 'get'
  })
}

// 创建画面
export function createScada(data) {
  return request({
    url: '/scada',
    method: 'post',
    data
  })
}

// 更新画面
export function updateScada(id, data) {
  return request({
    url: `/scada/${id}`,
    method: 'put',
    data
  })
}

// 删除画面
export function deleteScada(id) {
  return request({
    url: `/scada/${id}`,
    method: 'delete'
  })
}
