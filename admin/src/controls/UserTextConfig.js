export default {
  name: '文本',
  parts: [
    { name: 'default', label: '文本', styles: ['color', 'font_size', 'font_weight', 'text_align'] }
  ],
  inputs: [],
  outputs: [],
  events: [],
  properties: ['name', 'span', 'text'],
  span: 2,
  defaultStyles: {
    default: {
      color: '#333',
      font_size: '14px',
      font_weight: 'normal',
      text_align: 'left'
    }
  }
}
