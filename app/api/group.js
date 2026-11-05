import request from '@/utils/request'

export const searchGroups = (data) => {
  return request({
    url: '/group/search',
    method: 'POST',
    data
  })
}

export const getGroup = (id) => {
  return request({
    url: `/group/${id}`,
    method: 'GET'
  })
}

export const createGroup = (data) => {
  return request({
    url: '/group',
    method: 'POST',
    data
  })
}

export const updateGroup = (id, data) => {
  return request({
    url: `/group/${id}`,
    method: 'PUT',
    data
  })
}

export const deleteGroup = (id) => {
  return request({
    url: `/group/${id}`,
    method: 'DELETE'
  })
}