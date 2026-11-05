// 值类型转换工具函数

// 转为布尔值
export const toBool = (val) => {
  if (typeof val === 'boolean') return val
  if (typeof val === 'number') return val !== 0
  if (typeof val === 'string') return val === 'true' || val === '1' || val === 'on'
  return false
}

// 转为数值
export const toNum = (val, def = 0) => {
  if (typeof val === 'number') return val
  if (typeof val === 'string') {
    const n = parseFloat(val)
    return isNaN(n) ? def : n
  }
  if (typeof val === 'boolean') return val ? 1 : 0
  return def
}

// 转为字符串
export const toStr = (val) => {
  if (val === null || val === undefined) return ''
  return String(val)
}
