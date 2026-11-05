export default {
  name: '色温',
  parts: [
    { name: 'default', label: '轨道', styles: ['border_radius'] },
    { name: 'cool', label: '冷色端', styles: ['background_color'] },
    { name: 'warm', label: '暖色端', styles: ['background_color'] },
    { name: 'pointer', label: '指针', styles: ['background_color'] }
  ],
  inputs: [
    { name: 'value', label: '色温属性' }
  ],
  outputs: [
    { name: 'value', label: '色温值' }
  ],
  events: [
    { name: 'change', label: '调节事件' }
  ],
  properties: ['name', 'span', 'min', 'max', 'step', 'unit'],
  span: 2,
  min: 2700,
  max: 6500,
  step: 100,
  unit: 'K',
  defaultStyles: {
    default: {
      border_radius: '6px'
    },
    cool: {
      background_color: '#1E90FF'
    },
    warm: {
      background_color: '#FF8C00'
    },
    pointer: {
      background_color: '#ffffff'
    }
  }
}
