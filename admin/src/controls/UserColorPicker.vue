<template>
  <div class="control-color-picker">
    <div class="color-header">
      <span class="color-label">{{ control.name }}</span>
      <div class="color-value-box">
        <span class="color-hex">{{ rgbColor }}</span>
        <span class="color-block" :style="{ background: currentColor }"></span>
      </div>
    </div>

    <div class="color-controls">
      <div class="control-row">
        <span class="control-label">色相</span>
        <div class="hue-slider-wrapper">
          <input 
            type="range" 
            min="0" 
            max="360" 
            :value="hue"
            @input="updateHue"
            class="hue-slider"
          />
        </div>
      </div>
      
      <div class="control-row">
        <span class="control-label">亮度</span>
        <div class="brightness-slider-wrapper" :style="{ background: brightnessGradient }">
          <input 
            type="range" 
            min="0" 
            max="100" 
            :value="brightness"
            @input="updateBrightness"
            class="brightness-slider"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toStr } from './Control.js'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number],
    default: '#FF5733'
  }
})

const emit = defineEmits(['change', 'action'])

const hue = ref(0)
const brightness = ref(50)

const brightnessGradient = computed(() => {
  return `linear-gradient(to right, #000, hsl(${hue.value}, 100%, 50%), #fff)`
})

const currentColor = computed(() => {
  return `hsl(${hue.value}, 100%, ${brightness.value}%)`
})

const rgbColor = computed(() => {
  return hslToRgb(hue.value, 100, brightness.value).toUpperCase()
})

function hslToRgb(h, s, l) {
  s /= 100
  l /= 100
  const a = s * Math.min(l, 1 - l)
  const f = n => {
    const k = (n + h / 30) % 12
    const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1)
    return Math.round(255 * color).toString(16).padStart(2, '0')
  }
  return `#${f(0)}${f(8)}${f(4)}`
}

function parseColor(color) {
  if (!color || typeof color !== 'string') return
  
  color = color.toLowerCase()
  
  if (color.startsWith('hsl')) {
    const match = color.match(/hsl\((\d+),\s*(\d+)%,\s*(\d+)%\)/)
    if (match) {
      hue.value = parseInt(match[1])
      brightness.value = parseInt(match[3])
    }
  } else if (color.startsWith('#')) {
    let hex = color.replace('#', '')
    if (hex.length === 3) {
      hex = hex.split('').map(c => c + c).join('')
    }
    if (hex.length === 6) {
      const r = parseInt(hex.substring(0, 2), 16) / 255
      const g = parseInt(hex.substring(2, 4), 16) / 255
      const b = parseInt(hex.substring(4, 6), 16) / 255
      
      const max = Math.max(r, g, b)
      const min = Math.min(r, g, b)
      let h, s, l = (max + min) / 2
      
      if (max === min) {
        h = s = 0
      } else {
        const d = max - min
        s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
        switch (max) {
          case r: h = ((g - b) / d + (g < b ? 6 : 0)) / 6; break
          case g: h = ((b - r) / d + 2) / 6; break
          case b: h = ((r - g) / d + 4) / 6; break
        }
      }
      
      hue.value = Math.round(h * 360)
      brightness.value = Math.round(l * 100)
    }
  }
}

watch(() => props.value, (val) => {
  parseColor(toStr(val))
}, { immediate: true })

function updateHue(event) {
  hue.value = parseInt(event.target.value)
  emit('change', { type: 'color', value: rgbColor.value, control: props.control, bindKey: props.control.binds?.value })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'change',
      control: props.control,
      eventKey: 'change',
    })
  }
}

function updateBrightness(event) {
  brightness.value = parseInt(event.target.value)
  emit('change', { type: 'color', value: rgbColor.value, control: props.control, bindKey: props.control.binds?.value })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'change',
      control: props.control,
      eventKey: 'change',
    })
  }
}
</script>

<style lang="scss" scoped>
.control-color-picker {
  .color-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;

    .color-label {
      font-size: 14px;
      color: #333;
    }

    .color-value-box {
      display: flex;
      align-items: center;
      gap: 6px;

      .color-hex {
        font-size: 12px;
        color: #666;
        font-family: monospace;
      }

      .color-block {
        display: inline-block;
        width: 16px;
        height: 16px;
        border-radius: 4px;
        border: 1px solid #e0e0e0;
        box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        transition: background 0.2s;
      }
    }
  }

  .color-controls {
    .control-row {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 20px;

      &:last-child {
        margin-bottom: 0;
      }

      .control-label {
        font-size: 12px;
        color: #666;
        width: 50px;
      }

      .hue-slider-wrapper {
        flex: 1;
        padding: 6px 0;

        .hue-slider {
          width: 100%;
          height: 8px;
          -webkit-appearance: none;
          appearance: none;
          border-radius: 4px;
          background: linear-gradient(to right, 
            hsl(0, 100%, 50%),
            hsl(60, 100%, 50%),
            hsl(120, 100%, 50%),
            hsl(180, 100%, 50%),
            hsl(240, 100%, 50%),
            hsl(300, 100%, 50%),
            hsl(360, 100%, 50%)
          );
          outline: none;

          &::-webkit-slider-thumb {
            -webkit-appearance: none;
            width: 20px;
            height: 20px;
            background: #fff;
            border-radius: 50%;
            cursor: pointer;
            border: 2px solid #333;
            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
          }

          &::-moz-range-thumb {
            width: 20px;
            height: 20px;
            background: #fff;
            border-radius: 50%;
            cursor: pointer;
            border: 2px solid #333;
            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
          }
        }
      }

      .brightness-slider-wrapper {
        flex: 1;
        padding: 6px 0;
        border-radius: 4px;

        .brightness-slider {
          width: 100%;
          height: 8px;
          -webkit-appearance: none;
          appearance: none;
          border-radius: 4px;
          background: transparent;
          outline: none;

          &::-webkit-slider-thumb {
            -webkit-appearance: none;
            width: 20px;
            height: 20px;
            background: #fff;
            border-radius: 50%;
            cursor: pointer;
            border: 2px solid #333;
            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
          }

          &::-moz-range-thumb {
            width: 20px;
            height: 20px;
            background: #fff;
            border-radius: 50%;
            cursor: pointer;
            border: 2px solid #333;
            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
          }
        }
      }
    }
  }
}
</style>
