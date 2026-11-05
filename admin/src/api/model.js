import request from '@/utils/request'

export function getModel(id) {
  return request({
    url: `/model/${id}`,
    method: 'get'
  })
}

export function createModel(data) {
  return request({
    url: '/model',
    method: 'post',
    data
  })
}

export function updateModel(id, data) {
  return request({
    url: `/model/${id}`,
    method: 'put',
    data
  })
}



export function deleteModel(id) {
  return request({
    url: `/model/${id}`,
    method: 'delete'
  })
}

export function searchModels(data) {
  return request({
    url: '/model/search',
    method: 'post',
    data
  })
}
