<template>
  <view class="control-hue">
    <view class="hue-header">
      <text class="hue-label">{{ control.name }}</text>
      <view class="hue-indicator">
        <text class="hue-value">{{ currentValue }}°</text>
        <view class="hue-color-dot" :style="{ background: currentHueColor }"></view>
      </view>
    </view>
    <view class="hue-slider-container">
      <view class="hue-track">
        <view class="hue-gradient"></view>
        <view
          class="hue-pointer"
          :style="{ left: fillWidth + '%', background: currentHueColor }"
        ></view>
      </view>
      <slider
        :min="control.min || 0"
        :max="control.max || 360"
        :step="control.step || 1"
        :value="currentValue"
        activeColor="transparent"
        backgroundColor="transparent"
        block-size="20"
        @change="handleChange"
        @changing="handleChanging"
        class="hue-range"
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
    default: 0
  }
})

const emit = defineEmits(['change', 'action'])

const currentValue = ref(toNum(props.value, 0))

const min = computed(() => props.control.min || 0)
const max = computed(() => props.control.max || 360)

const fillWidth = computed(() => {
  return ((currentValue.value - min.value) / (max.value - min.value)) * 100
})

const currentHueColor = computed(() => {
  return `hsl(${currentValue.value}, 100%, 50%)`
})

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 0)
})

const handleChange = (event) => {
  currentValue.value = parseInt(event.detail.value)
  const bindKey = props.control.binds?.value
  emit('change', {
    type: 'hue',
    value: currentValue.value,
    controlId: props.control.id,
    bindKey
  })
  if (props.control.events?.change?.key) {
    emit('action', {
      type: 'hue',
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
.control-hue {
  .hue-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24rpx;

    .hue-label {
      font-size: 28rpx;
      color: #333;
    }

    .hue-indicator {
      display: flex;
      align-items: center;
      gap: 12rpx;

      .hue-value {
        font-size: 24rpx;
        color: #666;
      }

      .hue-color-dot {
        width: 32rpx;
        height: 32rpx;
        border-radius: 8rpx;
        border: 2rpx solid #e0e0e0;
      }
    }
  }

  .hue-slider-container {
    position: relative;
    height: 80rpx;
    display: flex;
    align-items: center;

    .hue-track {
      position: relative;
      width: 100%;
      height: 16rpx;
      border-radius: 8rpx;
      overflow: visible;

      .hue-gradient {
        position: absolute;
        left: 0;
        top: 0;
        width: 100%;
        height: 100%;
        background: linear-gradient(to right,
          hsl(0, 100%, 50%),
          hsl(60, 100%, 50%),
          hsl(120, 100%, 50%),
          hsl(180, 100%, 50%),
          hsl(240, 100%, 50%),
          hsl(300, 100%, 50%),
          hsl(360, 100%, 50%)
        );
        border-radius: 8rpx;
      }

      .hue-pointer {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        width: 40rpx;
        height: 40rpx;
        border-radius: 50%;
        border: 6rpx solid #fff;
        box-shadow: 0 4rpx 12rpx rgba(0,0,0,0.2);
        z-index: 1;
        pointer-events: none;
      }
    }

    .hue-range {
      position: absolute;
      width: 100%;
      height: 100%;
      z-index: 2;
      opacity: 0;
    }
  }
}
</style>
