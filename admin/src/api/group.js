import request from '@/utils/request'

export function searchGroups(data) {
  return request({
    url: '/group/search',
    method: 'post',
    data
  })
}

export function createGroup(data) {
  return request({
    url: '/group',
    method: 'post',
    data
  })
}

export function updateGroup(id, data) {
  return request({
    url: `/group/${id}`,
    method: 'put',
    data
  })
}

export function deleteGroup(id) {
  return request({
    url: `/group/${id}`,
    method: 'delete'
  })
}