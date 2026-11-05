import { useOEMStore } from '@/store/oem'

// 默认系统名称
const DEFAULT_SYSTEM_NAME = '物联小白'

// 默认 LOGO
const DEFAULT_LOGO = '/logo.jpg'

/**
 * 获取 LOGO（优先从 store/localStorage 获取）
 * @returns {string} LOGO 图片路径
 */
export const getLogo = () => {
  try {
    const store = useOEMStore()
    return store.getLogo
  } catch {
    try {
      const cached = localStorage.getItem('oem_config')
      if (cached) {
        return JSON.parse(cached).logo || DEFAULT_LOGO
      }
    } catch (e) {
      console.error('读取 OEM 配置失败', e)
    }
    return DEFAULT_LOGO
  }
}

/**
 * 获取系统名称（优先从 store/localStorage 获取）
 * @returns {string} 系统名称
 */
export const getSystemName = () => {
  try {
    const store = useOEMStore()
    return store.getSystemName
  } catch {
    try {
      const cached = localStorage.getItem('oem_config')
      if (cached) {
        return JSON.parse(cached).system_name || DEFAULT_SYSTEM_NAME
      }
    } catch (e) {
      console.error('读取 OEM 配置失败', e)
    }
    return DEFAULT_SYSTEM_NAME
  }
}

/**
 * 初始化 OEM 配置（在应用启动时调用）
 */
export const initOEM = async () => {
  try {
    const store = useOEMStore()
    await store.init()
  } catch (e) {
    console.error('初始化 OEM 配置失败', e)
  }
}
