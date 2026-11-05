<template>
  <view class="control-brightness">
    <view class="brightness-header">
      <text class="brightness-label">{{ control.name }}</text>
    </view>
    <text class="brightness-value">{{ currentValue }}{{ control.unit || '%' }}</text>
    <view class="brightness-slider">
      <view class="brightness-track">
        <view class="brightness-fill" :style="{ width: fillWidth + '%' }"></view>
        <view class="brightness-pointer" :style="{ left: fillWidth + '%' }"></view>
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
        class="brightness-range"
      />
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

const fillWidth = computed(() => {
  return ((currentValue.value - min.value) / (max.value - min.value)) * 100
})

const handleChange = (event) => {
  currentValue.value = parseInt(event.detail.value)
  const bindKey = props.control.binds?.value
  emit('change', {
    type: 'brightness',
    value: currentValue.value,
    controlId: props.control.id,
    bindKey
  })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'brightness',
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
.control-brightness {
  .brightness-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16rpx;

    .brightness-label {
      font-size: 28rpx;
      color: #333;
    }
  }

  .brightness-value {
    font-size: 36rpx;
    font-weight: 600;
    color: #FFD700;
    text-align: right;
    display: block;
    margin-bottom: 16rpx;
  }

  .brightness-slider {
    position: relative;
    height: 80rpx;
    display: flex;
    align-items: center;

    .brightness-track {
      position: relative;
      width: 100%;
      height: 16rpx;
      background: linear-gradient(to right, #333, #FFD700);
      border-radius: 8rpx;
      overflow: visible;

      .brightness-fill {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        background: linear-gradient(to right, #FFD700, #FFD700);
        border-radius: 8rpx;
        transition: width 0.1s;
      }

      .brightness-pointer {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 40rpx;
        height: 40rpx;
        background: #fff;
        border-radius: 50%;
        border: 6rpx solid #FFD700;
        box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.15);
        z-index: 1;
        pointer-events: none;
        transition: left 0.1s;
      }
    }

    .brightness-range {
      position: absolute;
      width: 100%;
      height: 100%;
      z-index: 2;
      opacity: 0;
    }
  }
}
</style>
