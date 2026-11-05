import request from '@/utils/request'

// 安装向导接口只存在于安装模式后端（默认 8080）中，
// 因此统一使用同源相对路径 /api，避免开发环境绝对地址指向正常模式后端导致 404
const INSTALL_BASE_URL = '/api'

// 测试关系数据库连接
export function testDB(data) {
  return request({
    baseURL: INSTALL_BASE_URL,
    url: '/setup/test-db',
    method: 'post',
    data
  })
}

// 测试 Redis 连通性
export function testRedis(data) {
  return request({
    baseURL: INSTALL_BASE_URL,
    url: '/setup/test-redis',
    method: 'post',
    data
  })
}

// 测试 InfluxDB 连接
export function testInflux(data) {
  return request({
    baseURL: INSTALL_BASE_URL,
    url: '/setup/test-influx',
    method: 'post',
    data
  })
}

// 保存安装配置并初始化数据库
export function saveSetup(data) {
  return request({
    baseURL: INSTALL_BASE_URL,
    url: '/setup/save',
    method: 'post',
    data
  })
}
