export default {
  name: '亮度',
  parts: [
    { name: 'default', label: '亮度条', styles: ['background_color', 'color', 'border_radius'] }
  ],
  inputs: [
    { name: 'value', label: '亮度属性' }
  ],
  outputs: [
    { name: 'value', label: '亮度值' }
  ],
  events: [
    { name: 'change', label: '调节事件' }
  ],
  properties: ['name', 'span', 'min', 'max', 'step', 'unit'],
  span: 2,
  min: 0,
  max: 100,
  step: 1,
  unit: '%',
  defaultStyles: {
    default: {
      background_color: '#FFD700',
      color: '#ffffff',
      border_radius: '8px'
    }
  }
}
