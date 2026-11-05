export default {
  name: '开关',
  parts: [
    { name: 'default', label: '开关', styles: ['background_color', 'color', 'border_radius'] },
    { name: 'icon', label: '图标', styles: ['icon', 'color', 'font_size'] },
    { name: 'thumb', label: '滑块', styles: ['background_color', 'border_radius'] }
  ],
  inputs: [
    { name: 'value', label: '状态属性' }
  ],
  outputs: [
    { name: 'value', label: '开关值' }
  ],
  events: [
    { name: 'toggle', label: '切换事件' }
  ],
  properties: ['name', 'span'],
  span: 1,
  defaultStyles: {
    default: {
      background_color: '#5BA3E8',
      color: '#ffffff',
      border_radius: '20px'
    },
    icon: {
      icon: '',
      color: '#999',
      font_size: '20px'
    },
    thumb: {
      background_color: '#ffffff',
      border_radius: '50%'
    }
  }
}
