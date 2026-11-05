import request from '@/utils/request'

// 分页搜索定时任务
export function searchTasks(data) {
  return request({
    url: '/task/search',
    method: 'post',
    data
  })
}

// 创建定时任务
export function createTask(data) {
  return request({
    url: '/task',
    method: 'post',
    data
  })
}

// 更新定时任务
export function updateTask(id, data) {
  return request({
    url: `/task/${id}`,
    method: 'put',
    data
  })
}

// 删除定时任务
export function deleteTask(id) {
  return request({
    url: `/task/${id}`,
    method: 'delete'
  })
}

// 启用/停用定时任务
export function switchTask(id, disabled) {
  return request({
    url: `/task/${id}/switch`,
    method: 'put',
    data: { disabled }
  })
}
