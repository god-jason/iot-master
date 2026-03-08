import {SmartField} from '../lib/smart-editor/smart-editor.component';
import {EChartsOption} from 'echarts';
import {SmartRequestService} from '../lib/smart-request.service';
import {SmartInfoItem} from '../lib/smart-info/smart-info.component';
import {ParamJoin, ParamSearch, SmartAction, SmartTableColumn, SmartTableOperator} from '../lib/smart-table/smart-table.component';

export type PageContent = Content & (
  BlankContent |
  TableContent |
  ImportContent |
  ExportContent |
  FormContent |
  InfoContent |
  ChartContent |
  MarkdownContent |
  StatisticContent |
  AmapContent)

export interface BlankContent {
  template?: 'blank' | ''
}

export interface Content {
  //子页面
  page?: string

  title?: string

  //初始化数据
  data?: any

  //数据接口
  load_api?: string
  load_success?: string | Function | ((data: any) => any)

  auto_refresh?: number

  //作为子页面时的参数（无用）
  params?: any
  params_func?: string | Function | ((data: any) => any)

//工具栏
  toolbar?: SmartField[]
//占用宽度 总数24
  span: string | number | null
  //高度
  height?: number | string

  //挂载
  mount?: string | Function | (() => void)
  //卸载
  unmount?: string | Function | (() => void)

  //注册成员
  methods?: { [key: string]: (string | Function | (() => any) | string[]) }

  //子页面
  children?: ChildPage[]

  //标签式子页面
  tabs?: TabPage[]
}

export interface ChildPage {
  page?: string
  span?: number
  content?: PageContent
  params?: any
  params_func?: string | Function | ((data: any) => any)
}

export interface TabPage {
  title?: string;
  page?: string
  content?: PageContent
  params?: any
  params_func?: string | Function | ((data: any) => any)
}


export interface AmapContent {
  template: 'amap'

  type: 'line' | 'point' | 'cluster' | 'animation'
  key?: string
  secret?: string
  style?: string
  zoom?: number
  city?: number
}

export interface ChartContent {
  template: 'chart'
  type: 'line' | 'bar' | 'pie' | 'gauge' | 'radar'
  //title: string;
  dark?: boolean
  theme?: string
  height?: number
  legend?: boolean
  tooltip?: boolean
  time?: boolean
  radar?: { [key: string]: number }
  gauge?: { key?: string }
  options: EChartsOption
}

export interface FormContent {
  template: 'form'
  fields: SmartField[]

  submit_api?: string
  submit?: string | Function | ((data: any) => Promise<any>)
  submit_success?: string | Function | ((data: any) => any)
}

export interface InfoContent {
  template: 'info'
  items: SmartInfoItem[]
}

export interface MarkdownContent {
  template: 'markdown'
  src?: string
}

export interface TableContent {
  template: 'table'
  columns: SmartTableColumn[]
  operators: SmartTableOperator[]
  keywords?: string[]
  joins?: ParamJoin[] //表关联

  search_api?: string
  search?: string | Function | ((event: ParamSearch, request: SmartRequestService) => Promise<any>)
}


export interface StatisticContent {
  template: 'statistic'
  items: Statistic[]
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
  columns: SmartTableColumn[],

  submit_api?: string
  submit?: string | Function | ((data: any) => Promise<any>)

  finish?: string | Function | ((data: any) => any)
}

export interface ExportContent {
  template: 'export'
  columns: SmartTableColumn[],

  search_api?: string
  search?: string | Function | ((event: ParamSearch, request: SmartRequestService) => Promise<any>)

  finish?: string | Function | ((data: any) => any)
}
