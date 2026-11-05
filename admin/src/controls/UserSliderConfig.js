export default {
  name: '滑块',
  parts: [
    { name: 'default', label: '轨道', styles: ['background_color', 'border_radius'] },
    { name: 'pointer', label: '指针', styles: ['background_color', 'width', 'height'] }
  ],
  inputs: [
    { name: 'value', label: '数值属性' }
  ],
  outputs: [
    { name: 'value', label: '数值' }
  ],
  events: [
    { name: 'change', label: '调节事件' }
  ],
  properties: ['name', 'span', 'min', 'max', 'step', 'unit'],
  span: 2,
  min: 0,
  max: 100,
  step: 1,
  unit: '',
  defaultStyles: {
    default: {
      background_color: '#e0e0e0',
      border_radius: '6px'
    },
    pointer: {
      background_color: '#5BA3E8',
      width: '20px',
      height: '20px'
    }
  }
}
