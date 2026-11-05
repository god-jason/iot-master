import request from '@/utils/request'

export function listAccounts(data) {
  return request({
    url: '/account/search',
    method: 'post',
    data
  })
}

export function getAccount(id) {
  return request({
    url: `/account/${id}`,
    method: 'get'
  })
}

export function createAccount(data) {
  return request({
    url: '/account',
    method: 'post',
    data
  })
}

export function updateAccount(id, data) {
  return request({
    url: `/account/${id}`,
    method: 'put',
    data
  })
}