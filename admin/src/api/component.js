import request from '@/utils/request'

// 分页搜索组件库列表
export function searchComponents(data) {
  return request({
    url: '/component/search',
    method: 'post',
    data
  })
}

// 获取组件详情
export function getComponent(id) {
  return request({
    url: `/component/${id}`,
    method: 'get'
  })
}

// 创建组件
export function createComponent(data) {
  return request({
    url: '/component',
    method: 'post',
    data
  })
}

// 更新组件
export function updateComponent(id, data) {
  return request({
    url: `/component/${id}`,
    method: 'put',
    data
  })
}

// 删除组件
export function deleteComponent(id) {
  return request({
    url: `/component/${id}`,
    method: 'delete'
  })
}
