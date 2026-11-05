import request from './request'

// 本地存储的缓存 key
const STORAGE_KEY = 'oem_config'

// 默认 OEM 配置（兜底值）
const DEFAULT_OEM = {
  logo: '',
  system_name: '物联小白',
  company: '',
  icp: ''
}

/**
 * 从本地存储读取缓存的 OEM 配置
 * @returns {Object} OEM 配置对象
 */
export const loadOEMFromStorage = () => {
  try {
    const cached = uni.getStorageSync(STORAGE_KEY)
    if (cached) {
      return typeof cached === 'string' ? JSON.parse(cached) : cached
    }
  } catch (e) {
    console.error('读取 OEM 缓存失败', e)
  }
  return { ...DEFAULT_OEM }
}

/**
 * 将 OEM 配置写入本地存储缓存
 * @param {Object} data OEM 配置对象
 */
export const saveOEMToStorage = (data) => {
  try {
    uni.setStorageSync(STORAGE_KEY, data)
  } catch (e) {
    console.error('写入 OEM 缓存失败', e)
  }
}

/**
 * 从后端接口获取最新 OEM 配置
 * @returns {Promise<Object>} OEM 配置对象
 */
export const fetchOEM = async () => {
  try {
    const res = await request({
      url: '/oem',
      method: 'GET',
      _skipAuthRedirect: true // OEM 接口在登录页也需要，跳过 401 自动跳转
    })
    if (res && res.data) {
      const data = res.data
      const oemData = {
        logo: data.logo || '',
        system_name: data.system_name || '',
        company: data.company || '',
        icp: data.icp || ''
      }
      saveOEMToStorage(oemData)
      return oemData
    }
  } catch (e) {
    console.error('获取 OEM 配置失败', e)
  }
  return { ...DEFAULT_OEM }
}

/**
 * 初始化 OEM 配置（在应用启动时调用）
 * 先读缓存快速展示，再异步拉取最新配置
 */
export const initOEM = async () => {
  // 先加载缓存
  loadOEMFromStorage()
  // 异步拉取最新
  await fetchOEM()
  // 更新所有页面的导航栏标题为系统名称
  const oem = getOEM()
  if (oem.system_name) {
    // #ifdef H5
    document.title = oem.system_name
    // #endif
    // #ifndef H5
    uni.setNavigationBarTitle({ title: oem.system_name })
    // #endif
  }
}

/**
 * 获取完整 OEM 配置（优先从缓存）
 * @returns {Object} OEM 配置对象
 */
export const getOEM = () => {
  return loadOEMFromStorage()
}

/**
 * 获取系统名称
 * @returns {string} 系统名称
 */
export const getSystemName = () => {
  const oem = getOEM()
  return oem.system_name || DEFAULT_OEM.system_name
}


