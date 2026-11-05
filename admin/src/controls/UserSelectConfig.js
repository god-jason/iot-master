export default {
  name: '模式选择',
  parts: [
    { name: 'default', label: '容器', styles: ['background_color', 'border_radius'] },
    { name: 'active', label: '选中项', styles: ['background_color', 'color', 'border_radius'] },
    { name: 'inactive', label: '未选中项', styles: ['color'] }
  ],
  inputs: [
    { name: 'value', label: '模式属性' }
  ],
  outputs: [
    { name: 'value', label: '模式值' }
  ],
  events: [
    { name: 'change', label: '切换事件' }
  ],
  properties: ['name', 'span'],
  span: 2,
  options: [
    { label: '自动', value: 'auto', icon: 'scene' },
    { label: '手动', value: 'manual', icon: 'illuminance' }
  ],
  defaultStyles: {
    default: {
      background_color: '#f5f5f5',
      border_radius: '8px'
    },
    active: {
      background_color: '#5BA3E8',
      color: '#ffffff',
      border_radius: '6px'
    },
    inactive: {
      color: '#666666'
    }
  }
}
