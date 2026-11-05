<template>
  <view class="control-number" :style="cardStyle">
    <view class="number-header">
      <text class="number-label">{{ control.name }}</text>
      <Icon 
        v-if="iconStyle.icon" 
        :name="iconStyle.icon" 
        :color="iconStyle.color"
        :size="iconStyle.font_size"
        class="number-icon" 
      />
    </view>
    <view class="number-value-wrap">
      <text class="number-value" :style="valueStyle">{{ displayValue }}</text>
      <text class="number-unit" :style="unitStyle">{{ control.unit || '' }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toNum } from './User.js'
import Icon from './Icon.vue'

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

const currentValue = ref(toNum(props.value, 0))

const iconStyle = computed(() => {
  const style = props.control.styles?.icon || {}
  return {
    icon: style.icon || '',
    color: style.color || '#999',
    font_size: style.font_size || '40rpx'
  }
})

const displayValue = computed(() => {
  const precision = props.control.precision != null ? props.control.precision : 0
  return Number(currentValue.value).toFixed(precision)
})

const cardStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    backgroundColor: style.background_color || '#f5f5f5',
    borderRadius: style.border_radius || '24rpx'
  }
})

const valueStyle = computed(() => {
  const style = props.control.styles?.text || {}
  return {
    color: style.color || '#333',
    fontSize: style.font_size || '56rpx'
  }
})

const unitStyle = computed(() => {
  const style = props.control.styles?.text || {}
  return {
    color: style.color ? style.color + '99' : '#999',
    fontSize: style.font_size ? `calc(${style.font_size} * 0.5)` : '28rpx'
  }
})

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 0)
})
</script>

<style lang="scss" scoped>
.control-number {
  padding: 32rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100%;
  box-sizing: border-box;

  .number-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .number-label {
      font-size: 26rpx;
      color: #999;
    }

    .number-icon {
      font-size: 40rpx;
    }
  }

  .number-value-wrap {
    display: flex;
    align-items: baseline;
    margin-top: 16rpx;

    .number-value {
      font-weight: 600;
      line-height: 1.2;
    }

    .number-unit {
      margin-left: 8rpx;
      font-weight: 500;
    }
  }
}
</style>
