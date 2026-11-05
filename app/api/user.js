import request from '@/utils/request'

export const getUserInfo = () => {
  return request({
    url: '/me',
    method: 'GET'
  })
}

export const updateUserInfo = (data) => {
  return request({
    url: '/me',
    method: 'PUT',
    data
  })
}

export const changePassword = (data) => {
  return request({
    url: '/password',
    method: 'POST',
    data
  })
}
