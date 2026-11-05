<template>
  <view class="control-slider">
    <view class="slider-header">
      <text class="slider-label">{{ control.name }}</text>
      <text class="slider-value">{{ currentValue }}{{ control.unit || '' }}</text>
    </view>
    <view class="slider-container">
      <view class="slider-track">
        <view class="slider-fill" :style="{ width: fillWidth + '%', background: trackColor }"></view>
        <view class="slider-thumb" :style="{ left: fillWidth + '%', borderColor: trackColor }"></view>
      </view>
      <slider 
        :min="control.min || 0" 
        :max="control.max || 100" 
        :step="control.step || 1"
        :value="currentValue"
        activeColor="transparent"
        backgroundColor="transparent"
        block-size="24"
        @change="handleChange"
        @changing="handleChanging"
        class="slider-range"
      />
    </view>
    <view class="slider-range-labels">
      <text>{{ control.min || 0 }}</text>
      <text>{{ control.max || 100 }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toNum } from './User.js'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number, Boolean],
    default: 50
  }
})

const emit = defineEmits(['change', 'action'])

const currentValue = ref(toNum(props.value, 50))

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 50)
})

const min = computed(() => props.control.min || 0)
const max = computed(() => props.control.max || 100)

const trackColor = computed(() => {
  const style = props.control.styles?.default || {}
  return style.color || '#5BA3E8'
})

const fillWidth = computed(() => {
  return ((currentValue.value - min.value) / (max.value - min.value)) * 100
})

const handleChange = (event) => {
  currentValue.value = parseInt(event.detail.value)
  const bindKey = props.control.binds?.value
  emit('change', {
    type: 'slider',
    value: currentValue.value,
    controlId: props.control.id,
    bindKey
  })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'slider',
      controlId: props.control.id,
      eventKey: 'change'
    })
  }
}

const handleChanging = (event) => {
  currentValue.value = parseInt(event.detail.value)
}
</script>

<style lang="scss" scoped>
.control-slider {
  .slider-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24rpx;

    .slider-label {
      font-size: 28rpx;
      color: #333;
    }

    .slider-value {
      font-size: 28rpx;
      font-weight: 600;
      color: #5BA3E8;
    }
  }

  .slider-container {
    position: relative;
    height: 60rpx;
    display: flex;
    align-items: center;

    .slider-track {
      position: relative;
      width: 100%;
      height: 12rpx;
      background: #e0e0e0;
      border-radius: 6rpx;

      .slider-fill {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        border-radius: 6rpx;
        transition: width 0.1s;
      }

      .slider-thumb {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 36rpx;
        height: 36rpx;
        background: #fff;
        border-radius: 50%;
        border-width: 6rpx;
        border-style: solid;
        box-shadow: 0 4rpx 8rpx rgba(0,0,0,0.2);
        z-index: 1;
        pointer-events: none;
      }
    }

    .slider-range {
      position: absolute;
      width: 100%;
      height: 100%;
      z-index: 2;
      opacity: 0;
    }
  }

  .slider-range-labels {
    display: flex;
    justify-content: space-between;
    margin-top: 8rpx;
    font-size: 24rpx;
    color: #999;
  }
}
</style>
