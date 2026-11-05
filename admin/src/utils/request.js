import axios from 'axios'
import Cookies from 'js-cookie'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = Cookies.get('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.code === 401) {
      Cookies.remove('token')
      setTimeout(() => {
        router.replace('/login')
      }, 0)
      ElMessage.error(res.message || '登录已过期')
      return Promise.reject(new Error(res.message))
    }
    if (res.code !== 0 && res.code !== undefined) {
      ElMessage.error(res.message || JSON.stringify(res))
      return Promise.reject(new Error(res.message || JSON.stringify(res)))
    }
    return res
  },
  (error) => {
    const errMsg = error.response?.data?.message || error.response?.data?.error || error.message || '网络错误'
    if (error.response?.status === 401) {
      Cookies.remove('token')
      setTimeout(() => {
        router.replace('/login')
      }, 100)
    }
    ElMessage.error(errMsg)
    return Promise.reject(error)
  }
)

export default request
