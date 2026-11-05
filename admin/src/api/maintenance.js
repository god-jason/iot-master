import request from '@/utils/request'

// ========== 维保计划 ==========

// 计划分页搜索
export function searchMaintenancePlans(data) {
  return request({
    url: '/maintenance/plan/search',
    method: 'post',
    data
  })
}

export function getMaintenancePlan(id) {
  return request({
    url: `/maintenance/plan/${id}`,
    method: 'get'
  })
}

// 创建计划：{ name, scope_type, scope_id, cycle_type, cycle_value, remind_days, items, owner_id, next_due, disabled }
export function createMaintenancePlan(data) {
  return request({
    url: '/maintenance/plan',
    method: 'post',
    data
  })
}

export function updateMaintenancePlan(id, data) {
  return request({
    url: `/maintenance/plan/${id}`,
    method: 'put',
    data
  })
}

export function deleteMaintenancePlan(id) {
  return request({
    url: `/maintenance/plan/${id}`,
    method: 'delete'
  })
}

// 启用/停用计划
export function switchMaintenancePlan(id, disabled) {
  return request({
    url: `/maintenance/plan/${id}/switch`,
    method: 'put',
    data: { disabled }
  })
}

// 立即生成工单
export function generateMaintenancePlan(id) {
  return request({
    url: `/maintenance/plan/${id}/generate`,
    method: 'post'
  })
}

// ========== 维保工单 ==========

// 工单分页搜索
export function searchMaintenanceOrders(data) {
  return request({
    url: '/maintenance/order/search',
    method: 'post',
    data
  })
}

export function getMaintenanceOrder(id) {
  return request({
    url: `/maintenance/order/${id}`,
    method: 'get'
  })
}

// 创建工单：{ device_id, type, title, priority, assignee_id, due_at, remark }
export function createMaintenanceOrder(data) {
  return request({
    url: '/maintenance/order',
    method: 'post',
    data
  })
}

export function updateMaintenanceOrder(id, data) {
  return request({
    url: `/maintenance/order/${id}`,
    method: 'put',
    data
  })
}

export function deleteMaintenanceOrder(id) {
  return request({
    url: `/maintenance/order/${id}`,
    method: 'delete'
  })
}

// 指派负责人
export function assignMaintenanceOrder(id, assigneeId) {
  return request({
    url: `/maintenance/order/${id}/assign`,
    method: 'put',
    data: { assignee_id: assigneeId }
  })
}

// 开始处理
export function startMaintenanceOrder(id) {
  return request({
    url: `/maintenance/order/${id}/start`,
    method: 'put'
  })
}

// 完成工单：{ result, items, attachments, parts, cost, remark }
export function completeMaintenanceOrder(id, data) {
  return request({
    url: `/maintenance/order/${id}/complete`,
    method: 'put',
    data
  })
}

// 取消工单：{ remark }
export function cancelMaintenanceOrder(id, remark) {
  return request({
    url: `/maintenance/order/${id}/cancel`,
    method: 'put',
    data: { remark }
  })
}

// 导出工单 CSV（参数与搜索一致，无分页）
export function exportMaintenanceOrders(params) {
  return request({
    url: '/maintenance/order/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// ========== 维保统计 ==========

export function getMaintenanceStats() {
  return request({
    url: '/maintenance/stats/summary',
    method: 'get'
  })
}
