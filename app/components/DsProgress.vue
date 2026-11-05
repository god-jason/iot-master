<template>
  <view class="ds-progress">
    <view v-if="showText" class="ds-progress-header">
      <text class="ds-progress-label">{{ label }}</text>
      <text class="ds-progress-value">{{ displayPercent }}</text>
    </view>
    <view class="ds-progress-bar">
      <view 
        class="ds-progress-fill" 
        :class="[`ds-progress-${type}`]"
        :style="fillStyle"
      ></view>
      <view v-if="showPercent" class="ds-progress-text-overlay">{{ displayPercent }}</view>
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
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    default: 100
  },
  label: {
    type: String,
    default: ''
  },
  type: {
    type: String,
    default: 'primary',
    validator: (val) => ['primary', 'success', 'warning', 'danger'].includes(val)
  },
  showText: {
    type: Boolean,
    default: true
  },
  showPercent: {
    type: Boolean,
    default: false
  },
  strokeWidth: {
    type: Number,
    default: 12
  },
  color: {
    type: String,
    default: ''
  }
})

const percent = computed(() => {
  return Math.min(100, Math.max(0, (props.value / props.max) * 100))
})

const displayPercent = computed(() => {
  return `${Math.round(percent.value)}%`
})

const fillStyle = computed(() => ({
  width: `${percent.value}%`,
  height: `${props.strokeWidth}rpx`,
  background: props.color || ''
}))
</script>

<style lang="scss">
@import './variables.scss';

.ds-progress {
  .ds-progress-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: $spacing-xs;

    .ds-progress-label {
      font-size: $font-size-sm;
      color: $text-secondary;
    }

    .ds-progress-value {
      font-size: $font-size-sm;
      color: $text-primary;
      font-weight: 500;
    }
  }

  .ds-progress-bar {
    position: relative;
    height: 24rpx;
    background: $bg-secondary;
    border-radius: 12rpx;
    overflow: hidden;

    .ds-progress-fill {
      border-radius: 12rpx;
      transition: width 0.3s ease;

      &.ds-progress-primary {
        background: $primary-color;
      }

      &.ds-progress-success {
        background: $success-color;
      }

      &.ds-progress-warning {
        background: $warning-color;
      }

      &.ds-progress-danger {
        background: $danger-color;
      }
    }

    .ds-progress-text-overlay {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      font-size: $font-size-xs;
      color: #fff;
      font-weight: 500;
    }
  }
}
</style>