// 组态/大屏编辑器图元库定义与工厂函数
// 所有图元均基于 leafer-editor 画布引擎创建
// Path 类天然激活 PointsEditTool，所以折线/曲线/管道（改用 Path 画流动虚线）选中后可双击编辑顶点
import { Rect, Ellipse, Line, Text, Image, Group, Path } from 'leafer-editor'

// 画布背景框的唯一标识，序列化时需排除
export const CANVAS_FRAME_ID = '__canvas_frame__'

// 默认画布配置
export const DEFAULT_CANVAS = {
  width: 1280,
  height: 720,
  fill: '#ffffff'
}

// 生成图元唯一 ID
let idSeed = 1
const genId = (prefix) => `${prefix}_${Date.now().toString(36)}_${(idSeed++).toString(36)}`

// 图元分组常量（前端展示用 + 后端 category 字段值）
export const ELEMENT_CATEGORY = {
  BASIC: 'basic',       // 基础图元：矩形/圆形/直线/折线/曲线/文本
  CONTROL: 'control'    // 页面控件：图片/按钮/仪表盘/进度条/开关/管道/图表/音频/视频
}

// 图元库元数据，用于编辑器左侧面板分组展示
// icon 为内联 SVG，与画布图元形状对应
// category: basic / control —— 用于 Editor.vue 左侧面板折叠分组
export const ELEMENT_LIBRARY = [
  // ===== 基础图元 =====
  {
    type: 'rect', category: 'basic', name: '矩形',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="6" y="10" width="28" height="20" rx="2" fill="none" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'ellipse', category: 'basic', name: '圆形',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><circle cx="20" cy="20" r="12" fill="none" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'line', category: 'basic', name: '直线',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><line x1="6" y1="32" x2="34" y2="8" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'polyline', category: 'basic', name: '折线',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><polyline points="5,30 14,12 24,26 35,8" fill="none" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'curve', category: 'basic', name: '曲线',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><path d="M5 30 C 14 6, 26 34, 35 10" fill="none" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'text', category: 'basic', name: '文本',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><text x="20" y="27" text-anchor="middle" font-size="20" fill="currentColor">T</text></svg>'
  },
  // ===== 页面控件 =====
  {
    type: 'image', category: 'control', name: '图片',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="6" y="8" width="28" height="24" rx="2" fill="none" stroke="currentColor" stroke-width="2"/><circle cx="14" cy="16" r="2.5" fill="currentColor"/><path d="M8 28 L18 20 L24 25 L30 18 L34 22" fill="none" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'button', category: 'control', name: '按钮',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="6" y="14" width="28" height="12" rx="4" fill="none" stroke="currentColor" stroke-width="2"/><text x="20" y="23" text-anchor="middle" font-size="9" fill="currentColor">BUTTON</text></svg>'
  },
  {
    type: 'gauge', category: 'control', name: '仪表盘',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><path d="M8 28 A 14 14 0 1 1 32 28" fill="none" stroke="currentColor" stroke-width="2.5"/><line x1="20" y1="24" x2="27" y2="15" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'progress', category: 'control', name: '进度条',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="5" y="16" width="30" height="8" rx="4" fill="none" stroke="currentColor" stroke-width="2"/><rect x="7" y="18" width="16" height="4" rx="2" fill="currentColor"/></svg>'
  },
  {
    type: 'switch', category: 'control', name: '开关',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="6" y="13" width="28" height="14" rx="7" fill="none" stroke="currentColor" stroke-width="2"/><circle cx="27" cy="20" r="4.5" fill="currentColor"/></svg>'
  },
  {
    type: 'pipe', category: 'control', name: '管道',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><line x1="5" y1="20" x2="35" y2="20" stroke="currentColor" stroke-width="7" stroke-linecap="round" opacity="0.35"/><line x1="7" y1="20" x2="33" y2="20" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-dasharray="6 5"/></svg>'
  },
  {
    type: 'chart-line', category: 'control', name: '折线图',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><polyline points="6,30 14,20 22,25 34,10" fill="none" stroke="currentColor" stroke-width="2"/><line x1="6" y1="34" x2="34" y2="34" stroke="currentColor" stroke-width="1.5"/></svg>'
  },
  {
    type: 'chart-bar', category: 'control', name: '柱状图',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="8" y="18" width="6" height="16" fill="currentColor"/><rect x="17" y="10" width="6" height="24" fill="currentColor"/><rect x="26" y="22" width="6" height="12" fill="currentColor"/></svg>'
  },
  {
    type: 'chart-pie', category: 'control', name: '饼图',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><circle cx="20" cy="20" r="13" fill="none" stroke="currentColor" stroke-width="2"/><path d="M20 20 L20 7 A13 13 0 0 1 32 25 Z" fill="currentColor"/></svg>'
  },
  {
    type: 'chart-gauge', category: 'control', name: '仪表图',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><path d="M7 28 A 14 14 0 1 1 33 28" fill="none" stroke="currentColor" stroke-width="2.5"/><line x1="20" y1="25" x2="26" y2="14" stroke="currentColor" stroke-width="2"/><circle cx="20" cy="25" r="2" fill="currentColor"/></svg>'
  },
  {
    type: 'audio', category: 'control', name: '音频',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="8" y="14" width="24" height="12" rx="3" fill="none" stroke="currentColor" stroke-width="2"/><path d="M16 18 L16 26 M20 16 L20 28 M24 19 L24 25" stroke="currentColor" stroke-width="2"/></svg>'
  },
  {
    type: 'video', category: 'control', name: '视频',
    icon: '<svg viewBox="0 0 40 40" width="28" height="28"><rect x="6" y="10" width="28" height="20" rx="2" fill="none" stroke="currentColor" stroke-width="2"/><polygon points="17,16 17,24 26,20" fill="currentColor"/></svg>'
  }
]

// 按 category 分组后的图元数组，供 Editor.vue 左侧面板使用
export function getElementsByCategory(category) {
  return ELEMENT_LIBRARY.filter(el => el.category === category)
}

// 图元分组显示顺序
export const CATEGORY_ORDER = ['basic', 'control']

// 分组中文名称映射
export const CATEGORY_NAMES = {
  basic: '基础图元',
  control: '页面控件'
}

// ========== 构造函数（内部使用） ==========

function createGauge() {
  const group = new Group({ id: genId('gauge'), name: '仪表盘', editable: true })
  const bg = new Ellipse({
    name: 'gauge_bg', editable: true,
    x: 0, y: 0, width: 120, height: 120,
    stroke: '#e6e8eb', strokeWidth: 12,
    startAngle: 150, endAngle: 390
  })
  const value = new Ellipse({
    name: 'gauge_value', editable: true,
    x: 0, y: 0, width: 120, height: 120,
    stroke: '#00d9ff', strokeWidth: 12,
    startAngle: 150, endAngle: 150
  })
  const label = new Text({
    name: 'gauge_label', editable: true,
    x: 0, y: 42, width: 120, height: 36,
    text: '0', fontSize: 24, fontWeight: 'bold', fontFamily: 'Arial',
    fill: '#333333', textAlign: 'center'
  })
  group.addMany(bg, value, label)
  group.data = { component: true, binding: { min: 0, max: 100 } }
  return group
}

function createProgress() {
  const group = new Group({ id: genId('progress'), name: '进度条', editable: true })
  const bg = new Rect({
    name: 'progress_bg', editable: true,
    x: 0, y: 0, width: 180, height: 18,
    fill: '#e6e8eb', cornerRadius: 9
  })
  const value = new Rect({
    name: 'progress_value', editable: true,
    x: 0, y: 0, width: 0, height: 18,
    fill: '#00d9ff', cornerRadius: 9
  })
  const label = new Text({
    name: 'progress_label', editable: true,
    x: 0, y: 22, width: 180, height: 18,
    text: '0%', fontSize: 12, fontFamily: 'Arial',
    fill: '#666666', textAlign: 'center'
  })
  group.addMany(bg, value, label)
  group.data = { component: true, binding: { min: 0, max: 100 } }
  return group
}

function createSwitch() {
  const group = new Group({ id: genId('switch'), name: '开关', editable: true })
  const bg = new Rect({
    name: 'switch_bg', editable: true,
    x: 0, y: 0, width: 64, height: 32,
    fill: '#c0c4cc', cornerRadius: 16
  })
  const dot = new Ellipse({
    name: 'switch_dot', editable: true,
    x: 36, y: 4, width: 24, height: 24,
    fill: '#ffffff'
  })
  group.addMany(bg, dot)
  group.data = { component: true }
  return group
}

// 创建管道图元（管体 + 流动虚线 Path）
// 内部的 pipe_flow 从 Line 改为 Path 类，让选中管道后可双击编辑流动路径顶点
function createPipe() {
  const group = new Group({ id: genId('pipe'), name: '管道', editable: true })
  const body = new Rect({
    name: 'pipe_body', editable: true,
    x: 0, y: 0, width: 200, height: 14,
    fill: '#9aa0b0', cornerRadius: 7, opacity: 0.45
  })
  // 流动层：Line + points 模式，激活 PointsEditTool 拖拽顶点
  const flow = new Line({
    name: 'pipe_flow', editable: true,
    x: 0, y: 0,
    stroke: '#00d9ff', strokeWidth: 4,
    strokeCap: 'round',
    dashPattern: [10, 8],
    dashOffset: 0,
    points: [4, 7, 196, 7]  // 水平线段 2 个顶点
  })
  group.addMany(body, flow)
  group.data = {
    component: true,
    pipe: {
      flowColor: '#00d9ff',
      speed: 3,
      direction: 1,
      animated: true
    }
  }
  return group
}

// 按钮图元（Rect + Text）
function createButton() {
  const group = new Group({ id: genId('button'), name: '按钮', editable: true })
  const bg = new Rect({
    name: 'button_bg', editable: true,
    x: 0, y: 0, width: 120, height: 36,
    fill: '#409EFF', cornerRadius: 6
  })
  const label = new Text({
    name: 'button_label', editable: true,
    x: 0, y: 0, width: 120, height: 36,
    text: '按钮', fontSize: 14, fontFamily: 'Arial', fontWeight: 'normal',
    fill: '#ffffff', textAlign: 'center', verticalAlign: 'middle'
  })
  group.addMany(bg, label)
  group.data = { component: true }
  return group
}

// 音频图元（占位框 + 播放按钮）
function createAudio() {
  const group = new Group({ id: genId('audio'), name: '音频', editable: true })
  const bg = new Rect({
    name: 'audio_bg', editable: true,
    x: 0, y: 0, width: 200, height: 56,
    fill: '#f5f7fa', stroke: '#dcdfe6', strokeWidth: 1, cornerRadius: 6
  })
  const play = new Ellipse({
    name: 'audio_play', editable: true,
    x: 8, y: 16, width: 24, height: 24,
    fill: '#409EFF'
  })
  const label = new Text({
    name: 'audio_label', editable: true,
    x: 38, y: 16, width: 154, height: 24,
    text: '音频控件', fontSize: 12, fontFamily: 'Arial',
    fill: '#606266', textAlign: 'left'
  })
  group.addMany(bg, play, label)
  group.data = { component: true, media: { type: 'audio' } }
  return group
}

// 视频图元（占位框 + 播放按钮覆盖层）
function createVideo() {
  const group = new Group({ id: genId('video'), name: '视频', editable: true })
  const bg = new Rect({
    name: 'video_bg', editable: true,
    x: 0, y: 0, width: 320, height: 200,
    fill: '#1a1a2e', stroke: '#303133', strokeWidth: 1, cornerRadius: 4
  })
  const play = new Ellipse({
    name: 'video_play', editable: true,
    x: 148, y: 88, width: 24, height: 24,
    fill: '#ffffff', opacity: 0.9
  })
  group.addMany(bg, play)
  group.data = { component: true, media: { type: 'video' } }
  return group
}

// 判断图元是否为管道图元
export function isPipeElement(el) {
  return !!(el && el.data && el.data.pipe)
}

// 图表类型名称映射
export const CHART_TYPE_NAMES = {
  line: '折线图',
  bar: '柱状图',
  pie: '饼图',
  gauge: '仪表图'
}

// 数据源类型名称映射
export const DATA_SOURCE_NAMES = {
  static: '静态数据',
  history: '历史聚合'
}

function createChart(chartType) {
  const name = CHART_TYPE_NAMES[chartType] || '图表'
  const rect = new Rect({
    id: genId('chart'),
    name,
    x: 100, y: 100, width: 360, height: 260,
    fill: '#1a1a2e',
    stroke: '#00d9ff',
    strokeWidth: 1,
    cornerRadius: 6,
    opacity: 0.9
  })
  rect.data = {
    chart: {
      chartType,
      title: name,
      dataSource: {
        type: 'static',
        staticData: [],
        device_id: '',
        property: '',
        window: '1h',
        method: 'mean',
        limit: 24
      }
    }
  }
  return rect
}

export function isChartElement(el) {
  return !!(el && el.data && el.data.chart && el.data.chart.chartType)
}

// 图元工厂：按类型创建新图元，默认放置在画布 (100,100) 附近
// 关键：必须显式加 editable: true，否则 EditSelectHelper.findOne 会把图元过滤掉
export function createElement(type) {
  switch (type) {
    case 'rect':
      return new Rect({ id: genId('rect'), name: '矩形', editable: true, x: 100, y: 100, width: 140, height: 90, fill: '#409EFF', cornerRadius: 4 })
    case 'ellipse':
      return new Ellipse({ id: genId('ellipse'), name: '圆形', editable: true, x: 100, y: 100, width: 100, height: 100, fill: '#67C23A' })
    case 'line':
      return new Line({ id: genId('line'), name: '直线', editable: true, x: 100, y: 100, toPoint: { x: 160, y: 0 }, stroke: '#303133', strokeWidth: 3 })
    case 'polyline':
      // 折线：用 Line + points 模式，自动 isPointsMode=true → 激活 PointsEditTool 顶点拖拽
      return new Line({
        id: genId('polyline'), name: '折线', editable: true,
        x: 100, y: 100, stroke: '#303133', strokeWidth: 3,
        points: [0, 60, 40, 10, 90, 50, 150, 0]
      })
    case 'curve':
      // 曲线：用 Path + 三次贝塞尔(C)命令，支持 Photoshop 式锚点/控制手柄编辑
      // 首个锚点(0,60) → 控制点(20,20)(70,100) → 末锚点(150,0)
      return new Path({
        id: genId('curve'), name: '曲线', editable: true,
        x: 100, y: 100, stroke: '#303133', strokeWidth: 3,
        path: 'M 0 60 C 20 20 70 100 150 0'
      })
    case 'text':
      return new Text({ id: genId('text'), name: '文本', editable: true, x: 100, y: 100, text: '双击编辑文本', fontSize: 20, fontFamily: 'Arial', fill: '#303133' })
    case 'image':
      return new Image({ id: genId('image'), name: '图片', editable: true, x: 100, y: 100, width: 160, height: 120, cornerRadius: 4, fill: '#f0f2f5' })
    case 'button':
      return createButton()
    case 'gauge':
      return createGauge()
    case 'progress':
      return createProgress()
    case 'switch':
      return createSwitch()
    case 'pipe':
      return createPipe()
    case 'chart-line':
      return createChart('line')
    case 'chart-bar':
      return createChart('bar')
    case 'chart-pie':
      return createChart('pie')
    case 'chart-gauge':
      return createChart('gauge')
    case 'audio':
      return createAudio()
    case 'video':
      return createVideo()
    default:
      return null
  }
}

// ========== 图元类型元信息（属性面板 / 图层列表等用） ==========

const TAG_NAMES = {
  Rect: '矩形', Ellipse: '圆形', Line: '直线', Text: '文本',
  Image: '图片', Group: '组合', Path: '路径'
}
export function getElementTypeName(el) {
  if (!el) return ''
  return el.name || TAG_NAMES[el.tag] || el.tag
}

// 文本图元可选字体
export const FONT_FAMILIES = [
  'Arial', 'Helvetica', 'Times New Roman', 'SimSun', 'SimHei', 'KaiTi',
  'Microsoft YaHei', 'Tahoma', 'Verdana', 'Courier New', 'Georgia'
]

// 递归把图元及其所有子元素的 editable 设为 true
// Leafer 的 EditSelectHelper.findOne 用 leaf.editable 过滤，默认 editable = false
// 所以必须显式设 editable = true 才能让图元被选中/拖动
export function patchEditable(el) {
  if (el && el.editable !== 'single') el.editable = true
  if (el && el.children) {
    el.children.forEach(patchEditable)
  }
}
