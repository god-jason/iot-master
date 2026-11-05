export default {
  name: '色相',
  parts: [
    { name: 'default', label: '色相条', styles: ['border_radius'] },
    { name: 'pointer', label: '指针', styles: ['background_color', 'border_color'] }
  ],
  inputs: [
    { name: 'value', label: '色相属性' }
  ],
  outputs: [
    { name: 'value', label: '色相值' }
  ],
  events: [
    { name: 'change', label: '调节事件' }
  ],
  properties: ['name', 'span', 'min', 'max', 'step', 'unit'],
  span: 2,
  min: 0,
  max: 360,
  step: 1,
  unit: '',
  defaultStyles: {
    default: {
      border_radius: '6px'
    },
    pointer: {
      background_color: '#ffffff',
      border_color: '#333333'
    }
  }
}
