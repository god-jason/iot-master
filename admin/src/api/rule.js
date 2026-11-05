import request from '@/utils/request'

// 分页搜索规则
export function searchRules(data) {
  return request({
    url: '/rule/search',
    method: 'post',
    data
  })
}

// 创建规则
export function createRule(data) {
  return request({
    url: '/rule',
    method: 'post',
    data
  })
}

// 更新规则
export function updateRule(id, data) {
  return request({
    url: `/rule/${id}`,
    method: 'put',
    data
  })
}

// 删除规则
export function deleteRule(id) {
  return request({
    url: `/rule/${id}`,
    method: 'delete'
  })
}

// 启用/停用规则
export function switchRule(id, disabled) {
  return request({
    url: `/rule/${id}/switch`,
    method: 'put',
    data: { disabled }
  })
}

// 模拟数据试跑规则
export function testRule(id, values) {
  return request({
    url: `/rule/${id}/test`,
    method: 'post',
    data: { values }
  })
}

// 分页搜索规则执行日志
export function searchRuleLogs(data) {
  return request({
    url: '/rule/logs/search',
    method: 'post',
    data
  })
}
