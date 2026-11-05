import request from '@/utils/request'

// 搜索维保工单
export const searchMaintenanceOrders = (params = {}) => {
  const body = {
    page: params.page || 1,
    page_size: params.page_size || 20
  }

  if (params.keyword) {
    body.keyword = params.keyword
  }
  if (params.status) {
    body.status = params.status
  }
  if (params.overdue !== undefined && params.overdue !== null && params.overdue !== '') {
    body.overdue = params.overdue
  }
  if (params.assignee_id) {
    body.assignee_id = params.assignee_id
  }

  return request({
    url: '/maintenance/order/search',
    method: 'POST',
    data: body
  })
}

// 获取维保工单详情
export const getMaintenanceOrderDetail = (id) => {
  return request({
    url: `/maintenance/order/${id}`,
    method: 'GET'
  })
}

// 开始处理工单
export const startMaintenanceOrder = (id) => {
  return request({
    url: `/maintenance/order/${id}/start`,
    method: 'PUT'
  })
}

// 提交完成工单
export const completeMaintenanceOrder = (id, data) => {
  return request({
    url: `/maintenance/order/${id}/complete`,
    method: 'PUT',
    data
  })
}

// 取消工单
export const cancelMaintenanceOrder = (id, remark) => {
  return request({
    url: `/maintenance/order/${id}/cancel`,
    method: 'PUT',
    data: {
      remark: remark || ''
    }
  })
}
