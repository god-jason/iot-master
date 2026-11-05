export default {
  name: '按钮',
  parts: [
    { name: 'default', label: '按钮', styles: ['background_color', 'color', 'font_size', 'border_radius', 'width', 'height', 'padding'] },
    { name: 'icon', label: '图标', styles: ['icon', 'color', 'font_size'] }
  ],
  inputs: [
    { name: 'value', label: '数值属性' }
  ],
  outputs: [],
  events: [
    { name: 'click', label: '点击事件' }
  ],
  properties: ['name', 'span', 'confirm'],
  span: 1,
  confirm: false,
  defaultStyles: {
    default: {
      background_color: '#5BA3E8',
      color: '#ffffff',
      font_size: '16px',
      border_radius: '8px',
      width: '100%',
      height: '100%',
      padding: '16px 24px'
    },
    icon: {
      icon: '',
      color: '#ffffff',
      font_size: '18px'
    }
  }
}
