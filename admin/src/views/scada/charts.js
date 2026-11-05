// echarts 图表图元渲染模块
// 负责：
//   - 按图表类型（折线/柱状/饼/仪表）与数据构建 echarts option
//   - 历史聚合数据源拉取（调用设备历史接口）
//   - 预览页图表覆盖层管理（DOM 定位随画布缩放、echarts 实例生命周期）
import * as echarts from 'echarts'
import { getDeviceHistory } from '@/api/device'

// 深色主题配色（与大屏风格一致）
const CHART_COLORS = ['#00d9ff', '#36e2a2', '#ffd166', '#ef767a', '#8e7cf3', '#4dabf7']

// 时间格式化：历史数据点横轴标签
function formatTime(ts) {
  const d = new Date(ts)
  const pad = n => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// 构建 echarts option
// @param chart - 图表配置 {chartType, title, dataSource}
// @param points - 数据点 [{name, value}]
// @return echarts option 对象
export function buildOption(chart, points) {
  const base = {
    backgroundColor: 'transparent',
    color: CHART_COLORS,
    title: chart.title ? {
      text: chart.title,
      left: 'center',
      textStyle: { color: '#ffffff', fontSize: 14, fontWeight: 'normal' }
    } : undefined,
    tooltip: { trigger: 'item' }
  }

  switch (chart.chartType) {
    case 'line':
      return {
        ...base,
        tooltip: { trigger: 'axis' },
        grid: { left: 40, right: 16, top: chart.title ? 40 : 20, bottom: 28 },
        xAxis: {
          type: 'category',
          data: points.map(p => p.name),
          axisLine: { lineStyle: { color: '#5a5a7a' } },
          axisLabel: { color: '#a0a0b8', fontSize: 10 }
        },
        yAxis: {
          type: 'value',
          splitLine: { lineStyle: { color: 'rgba(255,255,255,0.08)' } },
          axisLabel: { color: '#a0a0b8', fontSize: 10 }
        },
        series: [{
          type: 'line',
          smooth: true,
          symbol: 'none',
          data: points.map(p => p.value),
          areaStyle: { opacity: 0.15 }
        }]
      }
    case 'bar':
      return {
        ...base,
        tooltip: { trigger: 'axis' },
        grid: { left: 40, right: 16, top: chart.title ? 40 : 20, bottom: 28 },
        xAxis: {
          type: 'category',
          data: points.map(p => p.name),
          axisLine: { lineStyle: { color: '#5a5a7a' } },
          axisLabel: { color: '#a0a0b8', fontSize: 10 }
        },
        yAxis: {
          type: 'value',
          splitLine: { lineStyle: { color: 'rgba(255,255,255,0.08)' } },
          axisLabel: { color: '#a0a0b8', fontSize: 10 }
        },
        series: [{
          type: 'bar',
          barMaxWidth: 24,
          data: points.map(p => p.value)
        }]
      }
    case 'pie':
      return {
        ...base,
        tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
        series: [{
          type: 'pie',
          radius: ['35%', '62%'],
          center: ['50%', chart.title ? '58%' : '50%'],
          label: { color: '#a0a0b8', fontSize: 10 },
          data: points.map(p => ({ name: p.name, value: p.value }))
        }]
      }
    case 'gauge': {
      const value = points.length ? Number(points[points.length - 1].value) || 0 : 0
      return {
        ...base,
        series: [{
          type: 'gauge',
          center: ['50%', chart.title ? '62%' : '55%'],
          radius: '85%',
          min: 0,
          max: 100,
          progress: { show: true, width: 12, itemStyle: { color: '#00d9ff' } },
          axisLine: { lineStyle: { width: 12, color: [[1, 'rgba(255,255,255,0.12)']] } },
          axisTick: { show: false },
          splitLine: { length: 8, lineStyle: { color: '#5a5a7a' } },
          axisLabel: { color: '#a0a0b8', fontSize: 10, distance: 16 },
          pointer: { length: '60%', width: 4, itemStyle: { color: '#00d9ff' } },
          detail: {
            valueAnimation: true,
            color: '#ffffff',
            fontSize: 22,
            offsetCenter: [0, '70%'],
            formatter: '{value}'
          },
          data: [{ value: Math.max(0, Math.min(100, value)) }]
        }]
      }
    }
    default:
      return base
  }
}

// 拉取历史聚合数据并转换为数据点
// @param dataSource - 数据源配置 {device_id, property, window, method, limit}
// @return [{name, value}] 数据点列表（按时间升序）
export async function fetchHistoryPoints(dataSource) {
  if (!dataSource.device_id || !dataSource.property) return []
  const limit = Number(dataSource.limit) || 24
  // 时间范围：按聚合窗口反推起点
  const windowMs = parseWindow(dataSource.window || '1h')
  const end = new Date()
  const start = new Date(end.getTime() - windowMs * limit)
  try {
    const res = await getDeviceHistory(dataSource.device_id, {
      start: start.toISOString(),
      end: end.toISOString(),
      fields: dataSource.property,
      window: dataSource.window || '1h',
      method: dataSource.method || 'mean'
    })
    const seriesData = res.data && res.data.data ? res.data.data[dataSource.property] : null
    if (!Array.isArray(seriesData)) return []
    return seriesData
      .slice(-limit)
      .map(p => ({ name: formatTime(p.time), value: Number(p.value) || 0 }))
  } catch (e) {
    console.error('fetch history failed', e)
    return []
  }
}

// 解析聚合窗口字符串为毫秒数（如 5m / 1h / 1d）
function parseWindow(window) {
  const match = /^(\d+)([mhd])$/.exec(window)
  if (!match) return 3600 * 1000
  const num = Number(match[1])
  const unit = match[2]
  if (unit === 'm') return num * 60 * 1000
  if (unit === 'h') return num * 3600 * 1000
  return num * 24 * 3600 * 1000
}

// ChartOverlayManager 图表覆盖层管理器
// 在预览容器上按图元位置叠加 DOM 容器并渲染 echarts 图表，
// 随画布缩放同步位置与尺寸。以图元对象引用作为键，避免 ID 冲突
export class ChartOverlayManager {
  // @param container - 预览画布容器（position: relative）
  constructor(container) {
    this.container = container
    this.items = new Map() // el -> {el, chart, dom, instance}
    this.scale = 1
    this.offsetX = 0
    this.offsetY = 0
  }

  // 添加图表图元的覆盖层
  // @param el - leafer 图元（含 data.chart 配置）
  addChart(el) {
    if (!el || this.items.has(el)) return
    const chart = el.data && el.data.chart
    if (!chart) return
    const dom = document.createElement('div')
    dom.style.position = 'absolute'
    dom.style.pointerEvents = 'none'
    this.container.appendChild(dom)
    const instance = echarts.init(dom)
    this.items.set(el, { el, chart, dom, instance })
    this.layoutOne(el)
    // 初始渲染（静态数据或空数据）
    const points = Array.isArray(chart.dataSource?.staticData) ? chart.dataSource.staticData : []
    instance.setOption(buildOption(chart, points))
  }

  // 更新指定图元的数据并重绘
  updateData(el, points) {
    const item = this.items.get(el)
    if (!item) return
    item.instance.setOption(buildOption(item.chart, points), { notMerge: false })
  }

  // 获取图表配置
  getChart(el) {
    const item = this.items.get(el)
    return item ? item.chart : null
  }

  // 遍历所有图表项
  forEach(fn) {
    this.items.forEach((item) => fn(item))
  }

  // 更新整体布局（画布缩放/平移后调用）
  // @param scale - 画布缩放比例
  // @param offsetX - 画布左上角在容器中的 X 偏移
  // @param offsetY - 画布左上角在容器中的 Y 偏移
  updateLayout(scale, offsetX, offsetY) {
    this.scale = scale
    this.offsetX = offsetX
    this.offsetY = offsetY
    this.items.forEach((_, el) => this.layoutOne(el))
  }

  // 定位单个图表覆盖层
  layoutOne(el) {
    const item = this.items.get(el)
    if (!item) return
    const { dom, instance } = item
    const x = this.offsetX + (el.x || 0) * this.scale
    const y = this.offsetY + (el.y || 0) * this.scale
    const w = (el.width || 100) * this.scale
    const h = (el.height || 100) * this.scale
    dom.style.left = x + 'px'
    dom.style.top = y + 'px'
    dom.style.width = w + 'px'
    dom.style.height = h + 'px'
    instance.resize({ width: w, height: h })
  }

  // 销毁所有图表实例与 DOM
  destroy() {
    this.items.forEach(item => {
      item.instance.dispose()
      item.dom.remove()
    })
    this.items.clear()
  }
}
