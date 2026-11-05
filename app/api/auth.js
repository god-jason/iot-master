import request from '@/utils/request'

export const login = (data) => {
  return request({
    url: '/login',
    method: 'POST',
    data
  })
}

export const getCurrentUser = () => {
  return request({
    url: '/me',
    method: 'GET'
  })
}

export const wechatCode2Session = (code) => {
  return request({
    url: '/weixin/code2session',
    method: 'POST',
    data: { code }
  })
}

export const wechatAuth = (data) => {
  return request({
    url: '/weixin/auth',
    method: 'POST',
    data
  })
}

export const getPhoneNumber = (code) => {
  return request({
    url: '/weixin/phone',
    method: 'POST',
    data: { code }
  })
}

export const wechatLogin = (data) => {
  return request({
    url: '/weixin/auth',
    method: 'POST',
    data
  })
}

export const logout = () => {
  uni.removeStorageSync('token')
  uni.removeStorageSync('user')
  uni.removeStorageSync('current_family_id')
  return Promise.resolve({ code: 200 })
}

export const submitLoginToken = (id, token) => {
  return request({
    url: '/login/submit',
    method: 'POST',
    data: { id, token }
  })
}
