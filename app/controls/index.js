import UserButton from './UserButton.vue'
import UserSwitch from './UserSwitch.vue'
import UserSlider from './UserSlider.vue'
import UserBrightness from './UserBrightness.vue'
import UserSelect from './UserSelect.vue'
import UserColorPicker from './UserColorPicker.vue'
import UserColorTemperature from './UserColorTemperature.vue'
import UserColorHue from './UserColorHue.vue'
import UserNumber from './UserNumber.vue'
import UserText from './UserText.vue'
import UserVideo from './UserVideo.vue'
import UserImage from './UserImage.vue'
import Icon from './Icon.vue'

export {
  UserButton,
  UserSwitch,
  UserSlider,
  UserBrightness,
  UserSelect,
  UserColorPicker,
  UserColorTemperature,
  UserColorHue,
  UserNumber,
  UserText,
  UserVideo,
  UserImage,
  Icon
}

export const getControlComponent = (type) => {
  const components = {
    'button': UserButton,
    'switch': UserSwitch,
    'slider': UserSlider,
    'brightness': UserBrightness,
    'mode': UserSelect,
    'color': UserColorPicker,
    'color-temperature': UserColorTemperature,
    'color-hue': UserColorHue,
    'gauge': UserGauge,
    'number': UserNumber,
    'text': UserText,
    'video': UserVideo,
    'image': UserImage
  }
  return components[type] || UserNumber
}

export const getControlConfig = (type) => {
  const configs = {
    'button': {
      params: [{ name: 'value', label: '数值属性' }],
      actions: [{ name: 'click', label: '点击事件' }],
      parts: [
        { name: 'default', label: '按钮', styles: ['background_color', 'color', 'font_size', 'border_radius', 'width', 'height', 'padding'] },
        { name: 'icon', label: '图标', styles: ['icon', 'color', 'font_size'] }
      ],
      properties: ['name', 'span'],
      span: 1,
      defaultStyles: {
        default: {
          background_color: '#5BA3E8',
          color: '#ffffff',
          font_size: '14px',
          border_radius: '8px',
          width: '120px',
          height: '48px',
          padding: '12px 24px'
        },
        icon: {
          icon: '',
          color: '#ffffff',
          font_size: '18px'
        }
      }
    },
    'switch': {
      params: [{ name: 'value', label: '开关状态' }],
      actions: [],
      parts: [
        { name: 'default', label: '开关', styles: ['background_color', 'color', 'border_radius'] },
        { name: 'icon', label: '图标', styles: ['icon', 'color', 'font_size'] },
        { name: 'thumb', label: '滑块', styles: ['background_color', 'border_radius'] }
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
    },
    'slider': {
      params: [{ name: 'value', label: '数值属性' }],
      actions: [],
      parts: [{ name: 'default', label: '滑块', styles: ['color'] }],
      properties: ['name', 'span', 'min', 'max', 'step', 'unit'],
      span: 2,
      min: 0,
      max: 100,
      step: 1,
      unit: '%',
      defaultStyles: {
        default: {
          color: '#5BA3E8'
        }
      }
    },
    'brightness': {
      params: [{ name: 'value', label: '亮度值' }],
      actions: [],
      parts: [{ name: 'default', label: '亮度', styles: ['background_color', 'color', 'border_radius'] }],
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
    },
    'mode': {
      params: [{ name: 'value', label: '当前模式' }],
      actions: [{ name: 'select', label: '选择事件' }],
      parts: [
        { name: 'default', label: '容器', styles: ['background_color', 'border_radius'] },
        { name: 'active', label: '选中项', styles: ['background_color', 'color', 'border_radius'] },
        { name: 'inactive', label: '未选中项', styles: ['color'] }
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
    },
    'color': {
      params: [{ name: 'value', label: '颜色值' }],
      actions: [],
      parts: [{ name: 'default', label: '颜色选择', styles: ['color'] }],
      properties: ['name', 'span'],
      span: 1,
      defaultStyles: {
        default: {
          color: '#FF5733'
        }
      }
    },
    'color-temperature': {
      params: [{ name: 'value', label: '色温值' }],
      actions: [],
      parts: [{ name: 'default', label: '色温', styles: ['color'] }],
      properties: ['name', 'span', 'min', 'max', 'step', 'unit'],
      span: 2,
      min: 2700,
      max: 6500,
      step: 100,
      unit: 'K',
      defaultStyles: {
        default: {
          color: '#FFD700'
        }
      }
    },
    'hue': {
      params: [{ name: 'value', label: '色相值' }],
      actions: [],
      parts: [{ name: 'default', label: '色相', styles: ['color'] }],
      properties: ['name', 'span', 'min', 'max', 'step', 'unit'],
      span: 2,
      min: 0,
      max: 360,
      step: 1,
      unit: '°',
      defaultStyles: {
        default: {
          color: '#FF5733'
        }
      }
    },
    'gauge': {
      params: [{ name: 'value', label: '数值属性' }],
      actions: [],
      parts: [{ name: 'default', label: '仪表盘', styles: ['color', 'background_color', 'line_width'] }],
      properties: ['name', 'span', 'min', 'max', 'unit'],
      span: 2,
      min: 0,
      max: 100,
      unit: '%',
      defaultStyles: {
        default: {
          color: '#5BA3E8',
          background_color: '#e0e0e0',
          line_width: '8'
        }
      }
    },
    'number': {
      params: [{ name: 'value', label: '数值属性' }],
      actions: [],
      parts: [
        { name: 'default', label: '卡片', styles: ['background_color', 'border_radius'] },
        { name: 'text', label: '数值', styles: ['color', 'font_size'] },
        { name: 'icon', label: '图标', styles: ['icon', 'color', 'font_size'] }
      ],
      properties: ['name', 'span', 'unit', 'precision'],
      span: 1,
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
    },
    'temperature': {
      params: [{ name: 'value', label: '温度值' }],
      actions: [],
      parts: [{ name: 'default', label: '温度', styles: ['color', 'background_color', 'border_radius'] }],
      properties: ['name', 'span', 'unit'],
      span: 1,
      unit: '°C',
      defaultStyles: {
        default: {
          color: '#FF5733',
          background_color: '#fff5f5',
          border_radius: '8px'
        }
      }
    },
    'text': {
      params: [],
      actions: [],
      parts: [{ name: 'default', label: '文本', styles: ['color', 'font_size', 'font_weight', 'text_align'] }],
      properties: ['name', 'span', 'text'],
      span: 1,
      defaultStyles: {
        default: {
          color: '#333',
          font_size: '28rpx',
          font_weight: '400',
          text_align: 'left'
        }
      }
    },
    'video': {
      params: [],
      actions: [],
      parts: [{ name: 'default', label: '视频', styles: [] }],
      properties: ['name', 'span', 'url'],
      span: 1,
      defaultStyles: {
        default: {}
      }
    },
    'image': {
      params: [],
      actions: [],
      parts: [{ name: 'default', label: '图片', styles: [] }],
      properties: ['name', 'span', 'images'],
      span: 1,
      defaultStyles: {
        default: {}
      }
    }
  }
  return configs[type] || configs.number
}
