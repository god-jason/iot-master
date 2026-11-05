import request from '@/utils/request'
import { currentFamily } from '@/utils/user'

export const getDeviceList = (params) => {
  const body = {
    filter: [],
    page: params.page || 1,
    page_size: params.size || 20
  }

  if (params.keyword) {
    body.keyword = params.keyword
  }

  let url = '/device/search'
  if (currentFamily.value?.id) {
    url += '?family_id=' + currentFamily.value.id
  } else {
    return {
      code: 400,
      message: '没有家庭'
    }
  }

  if (params.room_id) {
    url += (url.includes('?') ? '&' : '?') + 'room_id=' + params.room_id
  }

  return request({
    url,
    method: 'POST',
    data: body
  })
}

export const getDeviceDetail = (id) => {
  return request({
    url: `/device/${id}`,
    method: 'GET'
  })
}

export const createDevice = (data) => {
  return request({
    url: '/device',
    method: 'POST',
    data
  })
}

export const updateDevice = (id, data) => {
  return request({
    url: `/device/${id}`,
    method: 'PUT',
    data
  })
}

export const getSubDevices = (gatewayId) => {
  return request({
    url: `/device/${gatewayId}/sub-device/search`,
    method: 'POST',
    data: {
      page: 1,
      page_size: 100
    }
  })
}

export const getDeviceLogs = (id, params) => {
  const body = {
    filter: [{ field: 'device_id', operator: 'eq', value: id }],
    order_by: 'created',
    order: 'desc',
    page: params.page || 1,
    page_size: params.size || 20
  }

  return request({
    url: '/device-log/search',
    method: 'POST',
    data: body
  })
}

export const getDeviceEvents = (params) => {
  const body = {
    filter: [],
    order_by: 'created',
    order: 'desc',
    page: params.page || 1,
    page_size: params.size || 20
  }

  if (params.deviceId) {
    body.filter.push({ field: 'device_id', operator: 'eq', value: params.deviceId })
  }

  return request({
    url: '/device-event/search',
    method: 'POST',
    data: body
  })
}
