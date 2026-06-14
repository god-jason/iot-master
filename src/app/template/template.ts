import {SmartField} from '../lib/smart-editor/smart-editor.component';
import {EChartsOption} from 'echarts';
import {SmartRequestService} from '../lib/smart-request.service';
import {SmartInfoItem} from '../lib/smart-info/smart-info.component';
import {
    ParamJoin,
    ParamSearch,
    SmartAction,
    SmartTableColumn,
    SmartTableOperator
} from '../lib/smart-table/smart-table.component';
import {Theme} from '@acrodata/code-editor';
import { CardTheme } from '../lib/smart-card/smart-card.component';

export type PageContent = Content & (
  BlankContent |
  LogContent |
  ListContent |
  ImportContent |
  ExportContent |
  EditContent |
  DetailContent |
  ChartContent |
  MarkdownContent |
  TextContent |
  StatisticContent |
  AmapContent |
  ValueContent)

export interface BlankContent {
  template?: 'blank' | ''
}

export interface Content {
  //子页面
  page?: string

  //卡片样式
  theme?: CardTheme
  style: any
  icon?: string
  title?: string
  titleStyle?: { [key: string]: any }
  bodyStyle:  { [key: string]: any }

  //初始化数据
  data?: any

  //数据接口
  load_api?: string
  load_success?: string | Function | ((data: any) => any)

  auto_refresh?: number

  //作为子页面时的参数（无用）
  params?: any | Function | ((data: any) => any)

//工具栏
  toolbar?: SmartField[]
//占用宽度 总数24
  span: string | number | null
  //高度
  height?: number | string

  //挂载
  mount?: string | Function | (() => void)
  mounts?: string[]

  //卸载
  unmount?: string | Function | (() => void)

  //注册成员
  methods?: { [key: string]: (string | Function | (() => any) | string[]) }


  //浮层子页面
  overlay?: ChildPage

  //子页面
  children?: ChildPage[]

  //标签式子页面
  tabs?: TabPage[]
}

export interface ChildPage {
  page?: string
  span?: number
  content?: PageContent
  params?: any | Function | ((data: any) => any)
  push?: string | number | null
}

export interface TabPage {
  title?: string;
  page?: string
  content?: PageContent
  params?: any | Function | ((data: any) => any)
}


export interface AmapContent {
  template: 'amap'

  type: 'line' | 'polygon' | 'polygons' | 'point' | 'cluster' | 'animation'
  key?: string
  secret?: string
  mapStyle?: string
  zoom?: number
  city?: number
  satellite?: boolean

  drawable?: boolean
  full?: boolean //全尺寸

  ready?: string | Function | (() => void) //初始化完成回调
  click?: string | Function | ((e: any) => void) //地图点击事件回调
}


export interface ChartContent {
  template: 'chart'
  type: 'line' | 'bar' | 'pie' | 'gauge' | 'radar'
  //title: string;
  dark?: boolean
  chartTheme?: string
  height?: number
  legend?: boolean
  tooltip?: boolean
  time?: boolean
  radar?: { [key: string]: number }
  gauge?: { key?: string }
  options: EChartsOption
}

export interface EditContent {
  template: 'edit'
  fields: SmartField[]

  submit_api?: string
  submit?: string | Function | ((data: any) => Promise<any>)
  submit_success?: string | Function | ((data: any) => any)
}

export interface DetailContent {
  template: 'detail'
  fields: SmartInfoItem[]
}

export interface MarkdownContent {
  template: 'markdown'
  src?: string
}

export interface TextContent {
  template: 'text'
  text?: string
}

export interface LogContent {
  template: 'log'

}

export interface ListContent {
  template: 'list'
  fields: SmartTableColumn[]
  operators: SmartTableOperator[]
  batch: boolean

  keywords?: string[]
  joins?: ParamJoin[] //表关联

  search_api?: string
  search?: string | Function | ((event: ParamSearch, request: SmartRequestService) => Promise<any>)
}


export interface StatisticContent {
  template: 'statistic'
  fields: Statistic[]
}

export interface ValueContent {
  template: 'value'
  
  // 显示数值（优先使用）
  value?: number | string
  // 标签/标题
  label?: string
  // 图片URL
  image?: string
  // 图片尺寸
  imageSize?: string | number
  // 图片位置：top/bottom/left/right/background
  imagePosition?: 'top' | 'bottom' | 'left' | 'right' | 'background'
  // 字体大小
  fontSize?: number | string
  // 字体颜色
  color?: string
  // 字体粗细
  fontWeight?: string | number
  // 字体样式
  fontFamily?: string
  // 前缀
  prefix?: string
  // 后缀
  suffix?: string
  // 数值格式化
  format?: string
  // 小数位数
  decimals?: number
  // 背景样式
  background?: string
  // 对齐方式
  align?: 'left' | 'center' | 'right'
  // 卡片样式
  style?: any
  bodyStyle?: any
  // 数值变化回调
  onChange?: string | Function | ((oldValue: any, newValue: any) => void)
}

export interface Statistic {
  key: string
  label: string
  span?: number
  format?: string
  prefix?: string
  suffix?: string
  action?: SmartAction
}

export interface ImportContent {
  template: 'import'
  fields: SmartTableColumn[],

  submit_api?: string
  submit?: string | Function | ((data: any) => Promise<any>)

  finish?: string | Function | ((data: any) => any)
}

export interface ExportContent {
  template: 'export'
  fields: SmartTableColumn[],

  search_api?: string
  search?: string | Function | ((event: ParamSearch, request: SmartRequestService) => Promise<any>)

  finish?: string | Function | ((data: any) => any)
}
