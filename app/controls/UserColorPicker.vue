<template>
  <view class="control-color-picker">
    <view class="color-header">
      <text class="color-label">{{ control.name }}</text>
      <view class="color-indicator">
        <text class="color-value">{{ rgbColor }}</text>
        <view class="color-dot" :style="{ background: currentColor }"></view>
      </view>
    </view>
    <view class="color-controls">
      <view class="control-row">
        <text class="control-label">色相</text>
        <view class="slider-wrapper">
          <view class="hue-track">
            <view class="hue-gradient"></view>
            <view class="slider-thumb" :style="{ left: huePercent + '%' }"></view>
          </view>
          <slider
            :value="hue"
            :min="0"
            :max="360"
            @change="updateHue"
            @changing="updateHueChanging"
            activeColor="transparent"
            backgroundColor="transparent"
            block-size="20"
            block-color="transparent"
            class="slider-input"
          />
        </view>
      </view>
      <view class="control-row">
        <text class="control-label">亮度</text>
        <view class="slider-wrapper">
          <view class="brightness-track" :style="{ background: brightnessGradient }">
            <view class="slider-thumb" :style="{ left: brightness + '%' }"></view>
          </view>
          <slider
            :value="brightness"
            :min="0"
            :max="100"
            @change="updateBrightness"
            @changing="updateBrightnessChanging"
            activeColor="transparent"
            backgroundColor="transparent"
            block-size="20"
            block-color="transparent"
            class="slider-input"
          />
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toStr } from './User.js'

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

const huePercent = computed(() => {
  return (hue.value / 360) * 100
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
  hue.value = parseInt(event.detail.value)
  const bindKey = props.control.binds?.value
  emit('change', {
    type: 'color',
    value: rgbColor.value,
    controlId: props.control.id,
    bindKey
  })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'color',
      controlId: props.control.id,
      eventKey: 'change'
    })
  }
}

function updateHueChanging(event) {
  hue.value = parseInt(event.detail.value)
}

function updateBrightness(event) {
  brightness.value = parseInt(event.detail.value)
  const bindKey = props.control.binds?.value
  emit('change', {
    type: 'color',
    value: rgbColor.value,
    controlId: props.control.id,
    bindKey
  })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'color',
      controlId: props.control.id,
      eventKey: 'change'
    })
  }
}

function updateBrightnessChanging(event) {
  brightness.value = parseInt(event.detail.value)
}
</script>

<style lang="scss" scoped>
.control-color-picker {
  .color-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24rpx;

    .color-label {
      font-size: 28rpx;
      color: #333;
    }

    .color-indicator {
      display: flex;
      align-items: center;
      gap: 12rpx;

      .color-value {
        font-size: 24rpx;
        color: #666;
        font-family: monospace;
      }

      .color-dot {
        width: 32rpx;
        height: 32rpx;
        border-radius: 8rpx;
        border: 2rpx solid #e0e0e0;
      }
    }
  }

  .color-controls {
    .control-row {
      display: flex;
      align-items: center;
      gap: 16rpx;
      margin-bottom: 32rpx;

      &:last-child {
        margin-bottom: 0;
      }

      .control-label {
        font-size: 24rpx;
        color: #666;
        width: 80rpx;
      }

      .slider-wrapper {
        flex: 1;
        position: relative;
        height: 60rpx;
        display: flex;
        align-items: center;

        .hue-track,
        .brightness-track {
          position: relative;
          width: 100%;
          height: 16rpx;
          border-radius: 8rpx;
          overflow: visible;
        }

        .hue-track {
          .hue-gradient {
            position: absolute;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            border-radius: 8rpx;
            background: linear-gradient(to right,
              hsl(0, 100%, 50%),
              hsl(60, 100%, 50%),
              hsl(120, 100%, 50%),
              hsl(180, 100%, 50%),
              hsl(240, 100%, 50%),
              hsl(300, 100%, 50%),
              hsl(360, 100%, 50%)
            );
          }
        }

        .slider-thumb {
          position: absolute;
          top: 50%;
          transform: translate(-50%, -50%);
          width: 40rpx;
          height: 40rpx;
          background: #fff;
          border-radius: 50%;
          border: 4rpx solid #333;
          box-shadow: 0 4rpx 8rpx rgba(0,0,0,0.2);
          z-index: 1;
          pointer-events: none;
        }

        .slider-input {
          position: absolute;
          width: 100%;
          height: 100%;
          z-index: 2;
          opacity: 0;
        }
      }
    }
  }
}
</style>
