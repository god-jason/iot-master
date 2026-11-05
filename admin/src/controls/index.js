import UserButton from './UserButton.vue'
import UserSwitch from './UserSwitch.vue'
import UserSlider from './UserSlider.vue'
import UserBrightness from './UserBrightness.vue'
import UserSelect from './UserSelect.vue'
import UserColorPicker from './UserColorPicker.vue'
import UserNumber from './UserNumber.vue'
import UserColorTemperature from './UserColorTemperature.vue'
import UserColorHue from './UserColorHue.vue'
import UserText from './UserText.vue'
import UserVideo from './UserVideo.vue'
import UserImage from './UserImage.vue'

import userButtonConfig from './UserButtonConfig.js'
import userSwitchConfig from './UserSwitchConfig.js'
import userSliderConfig from './UserSliderConfig.js'
import userBrightnessConfig from './UserBrightnessConfig.js'
import userSelectConfig from './UserSelectConfig.js'
import userColorConfig from './UserColorPickerConfig.js'
import userColorTempConfig from './UserColorTemperatureConfig.js'
import userColorHueConfig from './UserColorHueConfig.js'
import userNumberConfig from './UserNumberConfig.js'
import userTextConfig from './UserTextConfig.js'
import userVideoConfig from './UserVideoConfig.js'
import userImageConfig from './UserImageConfig.js'

export { UserButton, UserSwitch, UserSlider, UserBrightness, UserSelect, UserColorPicker, UserNumber, UserColorTemperature, UserColorHue, UserText, UserVideo, UserImage }

const defaultConfig = {
  parts: [
    { name: 'default', label: '默认', styles: ['background_color', 'color', 'font_size', 'border_radius'] }
  ],
  inputs: [
    { name: 'value', label: '数值属性' }
  ],
  outputs: [],
  events: [],
  properties: ['name', 'span'],
  span: 1,
  defaultStyles: {
    default: {
      background_color: '#f5f5f5',
      color: '#333',
      font_size: '14px',
      border_radius: '8px'
    }
  }
}

const allConfigs = {
  'button': userButtonConfig,
  'switch': userSwitchConfig,
  'slider': userSliderConfig,
  'brightness': userBrightnessConfig,
  'mode': userSelectConfig,
  'color': userColorConfig,
  'color-temperature': userColorTempConfig,
  'color-hue': userColorHueConfig,
  'number': userNumberConfig,
  'text': userTextConfig,
  'video': userVideoConfig,
  'image': userImageConfig
}

export const controlTypes = Object.entries(allConfigs).map(([type, config]) => ({
  type,
  label: config.name || type,
  icon: ''
}))

export const getControlConfig = (type) => {
  return allConfigs[type] || allConfigs['number'] || defaultConfig
}

export const controlConfigs = allConfigs

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
    'number': UserNumber,
    'text': UserText,
    'video': UserVideo,
    'image': UserImage
  }
  return components[type] || UserNumber
}

export const createControl = (type) => {
  const config = getControlConfig(type) || defaultConfig
  
  const binds = {}
  const inputs = Array.isArray(config.inputs) ? config.inputs : []
  inputs.forEach(p => {
    binds[p.name] = ''
  })
  
  const events = {}
  const eventList = Array.isArray(config.events) ? config.events : []
  eventList.forEach(e => {
    events[e.name] = { type: '', key: '', params: {} }
  })
  
  const styles = {}
  const parts = Array.isArray(config.parts) ? config.parts : []
  parts.forEach(part => {
    if (config.defaultStyles && config.defaultStyles[part.name]) {
      styles[part.name] = { ...config.defaultStyles[part.name] }
    }
  })
  
  const newControl = {
    id: Date.now().toString(),
    type,
    name: config.name || '新控件',
    span: config.span || 1,
    binds,
    events,
    styles,
    options: config.options ? JSON.parse(JSON.stringify(config.options)) : [],
    min: config.min || 0,
    max: config.max || 100,
    step: config.step || 1,
    unit: config.unit || '',
    precision: config.precision != null ? config.precision : 0,
    url: config.url || '',
    images: config.images ? [...config.images] : [],
    text: ''
  }
  
  return newControl
}
