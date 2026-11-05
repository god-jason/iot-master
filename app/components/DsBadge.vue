<template>
  <view class="ds-badge-wrapper" :style="wrapperStyle">
    <slot></slot>
    <view 
      v-if="visible && (value || dot)" 
      class="ds-badge" 
      :class="[`ds-badge-${type}`, { 'ds-badge-dot': dot }]"
    >
      <text v-if="!dot" class="ds-badge-value">{{ displayValue }}</text>
    </view>
  </view>
</template>

<script setup>
import { computed } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  value: {
    type: [Number, String],
    default: 0
  },
  max: {
    type: Number,
    default: 99
  },
  dot: {
    type: Boolean,
    default: false
  },
  visible: {
    type: Boolean,
    default: true
  },
  type: {
    type: String,
    default: 'danger',
    validator: (val) => ['danger', 'success', 'warning', 'info'].includes(val)
  },
  position: {
    type: String,
    default: 'top-right',
    validator: (val) => ['top-left', 'top-right', 'bottom-left', 'bottom-right'].includes(val)
  }
})

const displayValue = computed(() => {
  const val = Number(props.value)
  if (val > props.max) return `${props.max}+`
  return val
})

const wrapperStyle = computed(() => {
  const positions = {
    'top-left': { '--badge-top': '-4rpx', '--badge-left': '-4rpx' },
    'top-right': { '--badge-top': '-4rpx', '--badge-right': '-4rpx' },
    'bottom-left': { '--badge-bottom': '-4rpx', '--badge-left': '-4rpx' },
    'bottom-right': { '--badge-bottom': '-4rpx', '--badge-right': '-4rpx' }
  }
  return positions[props.position] || positions['top-right']
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-badge-wrapper {
  position: relative;
  display: inline-flex;
}

.ds-badge {
  position: absolute;
  min-width: 32rpx;
  height: 32rpx;
  padding: 0 8rpx;
  border-radius: 16rpx;
  background: $danger-color;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 4rpx solid $bg-primary;

  &.ds-badge-top-left {
    top: var(--badge-top);
    left: var(--badge-left);
  }

  &.ds-badge-top-right {
    top: var(--badge-top);
    right: var(--badge-right);
  }

  &.ds-badge-bottom-left {
    bottom: var(--badge-bottom);
    left: var(--badge-left);
  }

  &.ds-badge-bottom-right {
    bottom: var(--badge-bottom);
    right: var(--badge-right);
  }

  &.ds-badge-dot {
    width: 24rpx;
    height: 24rpx;
    min-width: auto;
    border-radius: 50%;
    padding: 0;
  }

  &.ds-badge-success {
    background: $success-color;
  }

  &.ds-badge-warning {
    background: $warning-color;
  }

  &.ds-badge-info {
    background: $info-color;
  }

  .ds-badge-value {
    font-size: $font-size-xs;
    color: #fff;
    line-height: 1;
  }
}
</style>