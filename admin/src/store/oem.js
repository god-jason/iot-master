import { defineStore } from 'pinia'
import { getOEM } from '@/api/oem'

// 本地存储的缓存 key
const STORAGE_KEY = 'oem_config'

// 默认 OEM 配置（当后端接口不可用时使用的兜底值）
const DEFAULT_OEM = {
  logo: '/logo.jpg',
  system_name: '物联小白',
  company: '',
  icp: ''
}

export const useOEMStore = defineStore('oem', {
  state: () => ({
    logo: '',
    systemName: '',
    company: '',
    icp: '',
    loaded: false
  }),
  getters: {
    // 获取系统名称（兜底默认值）
    getSystemName: (state) => state.systemName || DEFAULT_OEM.system_name,
    // 获取 logo（兜底默认值）
    getLogo: (state) => state.logo || DEFAULT_OEM.logo
  },
  actions: {
    // 从本地存储加载缓存的 OEM 配置
    loadFromStorage() {
      try {
        const cached = localStorage.getItem(STORAGE_KEY)
        if (cached) {
          const data = JSON.parse(cached)
          this.logo = data.logo || ''
          this.systemName = data.system_name || ''
          this.company = data.company || ''
          this.icp = data.icp || ''
          this.loaded = true
          // 同步更新页面标题
          document.title = this.systemName || DEFAULT_OEM.system_name
        }
      } catch (e) {
        console.error('读取 OEM 缓存失败', e)
      }
    },

    // 从后端接口获取最新 OEM 配置
    async fetch() {
      try {
        const res = await getOEM()
        if (res && res.data) {
          const data = res.data
          this.logo = data.logo || ''
          this.systemName = data.system_name || ''
          this.company = data.company || ''
          this.icp = data.icp || ''
          this.loaded = true
          // 写入本地存储缓存
          localStorage.setItem(STORAGE_KEY, JSON.stringify({
            logo: this.logo,
            system_name: this.systemName,
            company: this.company,
            icp: this.icp
          }))
          // 更新页面标题
          document.title = this.systemName || DEFAULT_OEM.system_name
        }
      } catch (e) {
        console.error('获取 OEM 配置失败', e)
      }
    },

    // 初始化：先读缓存快速展示，再异步拉取最新配置
    async init() {
      this.loadFromStorage()
      await this.fetch()
    },

    // 清除缓存（登出时调用）
    clearCache() {
      localStorage.removeItem(STORAGE_KEY)
    }
  }
})
