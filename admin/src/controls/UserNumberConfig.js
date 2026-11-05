export default {
  name: '数值',
  parts: [
    { name: 'default', label: '卡片', styles: ['background_color', 'border_radius'] },
    { name: 'text', label: '数值', styles: ['color', 'font_size'] },
    { name: 'icon', label: '图标', styles: ['icon', 'color', 'font_size'] }
  ],
  inputs: [
    { name: 'value', label: '温度属性' }
  ],
  outputs: [],
  events: [],
  properties: ['name', 'span', 'unit', 'precision'],
  span: 1,
  unit: '°C',
  precision: 0,
  defaultStyles: {
    default: {
      background_color: '#f5f5f5',
      border_radius: '12px'
    },
    text: {
      color: '#5BA3E8',
      font_size: '24px'
    },
    icon: {
      icon: 'temperature',
      color: '#999999',
      font_size: '20px'
    }
  }
}
