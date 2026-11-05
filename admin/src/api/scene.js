import request from '@/utils/request'

// 分页搜索场景
export function searchScenes(data) {
  return request({
    url: '/scene/search',
    method: 'post',
    data
  })
}

// 创建场景
export function createScene(data) {
  return request({
    url: '/scene',
    method: 'post',
    data
  })
}

// 更新场景
export function updateScene(id, data) {
  return request({
    url: `/scene/${id}`,
    method: 'put',
    data
  })
}

// 删除场景
export function deleteScene(id) {
  return request({
    url: `/scene/${id}`,
    method: 'delete'
  })
}

// 启用/停用场景
export function switchScene(id, disabled) {
  return request({
    url: `/scene/${id}/switch`,
    method: 'put',
    data: { disabled }
  })
}

// 手动执行场景
export function executeScene(id) {
  return request({
    url: `/scene/${id}/execute`,
    method: 'post'
  })
}
