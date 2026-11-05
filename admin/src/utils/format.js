import dayjs from 'dayjs'

/**
 * 日期格式化工具
 * @param {string|Date|number} date - 日期字符串、Date对象或unix时间戳
 * @param {string} format - 格式字符串，默认 'YYYY-MM-DD HH:mm:ss'
 * @returns {string} 格式化后的日期字符串
 */
export function formatDateTime(date, format = 'YYYY-MM-DD HH:mm:ss') {
  if (!date) return ''

  const d = dayjs(date)
  if (!d.isValid()) return ''

  return d.format(format)
}

/**
 * Vue全局过滤器
 */
export function registerFilters(app) {
  app.config.globalProperties.$formatDateTime = formatDateTime
  app.directive('format-datetime', {
    mounted(el, binding) {
      const format = binding.arg || 'YYYY-MM-DD HH:mm:ss'
      el.textContent = formatDateTime(el.textContent, format)
    }
  })
}
