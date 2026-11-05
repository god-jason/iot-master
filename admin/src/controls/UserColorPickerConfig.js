export default {
  name: '颜色选择',
  parts: [
    { name: 'default', label: '预览', styles: ['border_radius'] },
    { name: 'picker', label: '色盘', styles: ['background_color'] }
  ],
  inputs: [
    { name: 'value', label: '颜色属性' }
  ],
  outputs: [
    { name: 'value', label: '颜色值' }
  ],
  events: [
    { name: 'change', label: '设置事件' }
  ],
  properties: ['name', 'span'],
  span: 2,
  defaultStyles: {
    default: {
      border_radius: '50%'
    },
    picker: {
      background_color: '#ffffff'
    }
  }
}
