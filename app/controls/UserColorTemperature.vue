<template>
  <view class="control-color-temperature">
    <view class="temp-header">
      <text class="temp-label">{{ control.name }}</text>
      <text class="temp-value">{{ currentValue }}{{ control.unit || 'K' }}</text>
    </view>
    <view class="temp-slider-container">
      <view class="temp-track">
        <view
          class="temp-fill"
          :style="{ width: fillWidth + '%', background: gradientBackground }"
        ></view>
        <view
          class="temp-pointer"
          :style="{ left: fillWidth + '%' }"
        ></view>
      </view>
      <slider
        :min="control.min || 2700"
        :max="control.max || 6500"
        :step="control.step || 100"
        :value="currentValue"
        activeColor="transparent"
        backgroundColor="transparent"
        block-size="20"
        @change="handleChange"
        @changing="handleChanging"
        class="temp-range"
      />
    </view>
    <view class="temp-labels">
      <text class="temp-cool">冷色</text>
      <text class="temp-warm">暖色</text>
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
    default: 4000
  }
})

const emit = defineEmits(['change', 'action'])

const currentValue = ref(toNum(props.value, 4000))

const min = computed(() => props.control.min || 2700)
const max = computed(() => props.control.max || 6500)

const fillWidth = computed(() => {
  return ((currentValue.value - min.value) / (max.value - min.value)) * 100
})

const gradientBackground = computed(() => {
  return 'linear-gradient(to right, #1E90FF, #FF8C00)'
})

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 4000)
})

const handleChange = (event) => {
  currentValue.value = parseInt(event.detail.value)
  const bindKey = props.control.binds?.value
  emit('change', {
    type: 'color-temperature',
    value: currentValue.value,
    controlId: props.control.id,
    bindKey
  })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'color-temperature',
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
.control-color-temperature {
  .temp-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24rpx;

    .temp-label {
      font-size: 28rpx;
      color: #333;
    }

    .temp-value {
      font-size: 24rpx;
      color: #666;
    }
  }

  .temp-slider-container {
    position: relative;
    height: 80rpx;
    display: flex;
    align-items: center;

    .temp-track {
      position: relative;
      width: 100%;
      height: 16rpx;
      background: #e0e0e0;
      border-radius: 8rpx;
      overflow: visible;

      .temp-fill {
        position: absolute;
        left: 0;
        top: 0;
        height: 100%;
        border-radius: 8rpx;
        transition: width 0.1s;
      }

      .temp-pointer {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 40rpx;
        height: 40rpx;
        background: #fff;
        border-radius: 50%;
        border: 6rpx solid #5BA3E8;
        box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.15);
        z-index: 1;
        pointer-events: none;
      }
    }

    .temp-range {
      position: absolute;
      width: 100%;
      height: 100%;
      z-index: 2;
      opacity: 0;
    }
  }

  .temp-labels {
    display: flex;
    justify-content: space-between;
    margin-top: 16rpx;

    .temp-cool {
      font-size: 24rpx;
      color: #1E90FF;
    }

    .temp-warm {
      font-size: 24rpx;
      color: #FF8C00;
    }
  }
}
</style>
