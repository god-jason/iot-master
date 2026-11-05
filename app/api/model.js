import request from '@/utils/request'

export const getModel = (id) => {
  return request({
    url: `/model/${id}`,
    method: 'GET'
  })
}

export const getModelList = (params) => {
  return request({
    url: '/model/search',
    method: 'POST',
    data: params
  })
}
