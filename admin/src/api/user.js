import request from '@/utils/request'

export function getMe() {
  return request({
    url: '/me',
    method: 'get'
  })
}

export function updateMe(data) {
  return request({
    url: '/me',
    method: 'put',
    data
  })
}

export function changePassword(data) {
  return request({
    url: '/password',
    method: 'post',
    data
  })
}
