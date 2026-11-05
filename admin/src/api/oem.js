import request from '@/utils/request'

// 获取 OEM 定制化配置（公开接口，无需登录）
export function getOEM() {
  return request({
    url: '/oem',
    method: 'GET',
    _skipAuthRedirect: true // OEM 接口在登录页也需要，跳过 401 自动跳转
  })
}
