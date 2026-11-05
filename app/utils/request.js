const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'

import { navigateToLogin } from './navigation.js'

const request = (options) => {
  const token = uni.getStorageSync('token')
  
  return new Promise((resolve, reject) => {
    uni.request({
      url: baseURL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : '',
        ...options.header
      },
      timeout: 30000,
      success: (res) => {
        const { statusCode, data } = res
        
        if (statusCode === 200) {
          if (data.code === 401) {
            uni.removeStorageSync('token')
            uni.removeStorageSync('user')
            uni.removeStorageSync('current_family_id')
            if (!options._skipAuthRedirect) {
              navigateToLogin()
            }
            reject(new Error(data.message || '登录已过期'))
          } else if (data.code !== 0 && data.code !== undefined) {
            const errMsg = data.message || JSON.stringify(data)
            uni.showToast({
              title: errMsg.length > 20 ? errMsg.substring(0, 20) + '...' : errMsg,
              icon: 'none',
              duration: 3000
            })
            reject(new Error(errMsg))
          } else {
            resolve(data)
          }
        } else {
          const errMsg = data?.message || data?.error || `HTTP ${statusCode}` || '网络错误'
          uni.showToast({
            title: errMsg.length > 20 ? errMsg.substring(0, 20) + '...' : errMsg,
            icon: 'error',
            duration: 3000
          })
          if (statusCode === 401) {
            uni.removeStorageSync('token')
            uni.removeStorageSync('user')
            uni.removeStorageSync('current_family_id')
            if (!options._skipAuthRedirect) {
              navigateToLogin()
            }
          }
          reject(new Error(errMsg))
        }
      },
      fail: (err) => {
        const errMsg = err.errMsg || JSON.stringify(err) || '请求失败'
        uni.showToast({
          title: errMsg.length > 20 ? errMsg.substring(0, 20) + '...' : errMsg,
          icon: 'none',
          duration: 3000
        })
        reject(err)
      }
    })
  })
}

export default request
