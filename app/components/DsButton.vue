<template>
  <view 
    class="ds-button" 
    :class="[`ds-button-${type}`, `ds-button-${size}`, { 'ds-loading': loading, 'ds-disabled': disabled }]"
    :style="buttonStyle"
    @click="handleClick"
  >
    <view v-if="loading" class="ds-loading-icon">
      <view class="ds-loading-spinner"></view>
    </view>
    <DsIcon v-else-if="icon" :name="icon" :size="iconSize" :color="iconColor" />
    <text class="ds-button-text">{{ text }}</text>
    <slot></slot>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import DsIcon from './DsIcon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  text: {
    type: String,
    default: ''
  },
  type: {
    type: String,
    default: 'primary',
    validator: (val) => ['primary', 'success', 'warning', 'danger', 'default', 'ghost'].includes(val)
  },
  size: {
    type: String,
    default: 'medium',
    validator: (val) => ['mini', 'small', 'medium', 'large'].includes(val)
  },
  icon: {
    type: String,
    default: ''
  },
  iconColor: {
    type: String,
    default: ''
  },
  loading: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  },
  round: {
    type: Boolean,
    default: false
  },
  block: {
    type: Boolean,
    default: false
  },
  bgColor: {
    type: String,
    default: ''
  },
  textColor: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['click'])

const iconSize = computed(() => {
  const sizeMap = {
    mini: 20,
    small: 24,
    medium: 28,
    large: 32
  }
  return sizeMap[props.size] || 28
})

const buttonStyle = computed(() => {
  const style = {}
  if (props.bgColor) style.background = props.bgColor
  if (props.textColor) style.color = props.textColor
  if (props.round) style.borderRadius = '9999rpx'
  if (props.block) style.width = '100%'
  return style
})

const handleClick = (e) => {
  if (props.loading || props.disabled) return
  emit('click', e)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  max-width: 100%;
  gap: $spacing-xs;
  padding: $spacing-sm $spacing-lg;
  border-radius: $radius-base;
  font-size: $font-size-base;
  font-weight: 500;
  transition: $transition-base;
  user-select: none;
  -webkit-tap-highlight-color: transparent;

  &.ds-button-primary {
    background: $primary-color;
    color: #fff;
    box-shadow: $shadow-sm;
  }

  &.ds-button-success {
    background: $success-color;
    color: #fff;
  }

  &.ds-button-warning {
    background: $warning-color;
    color: #fff;
  }

  &.ds-button-danger {
    background: $danger-color;
    color: #fff;
  }

  &.ds-button-default {
    background: $bg-tertiary;
    color: $text-primary;
  }

  &.ds-button-ghost {
    background: transparent;
    color: $primary-color;
    border: 2rpx solid $primary-color;
  }

  &.ds-button-mini {
    padding: $spacing-xs $spacing-sm;
    font-size: $font-size-xs;
  }

  &.ds-button-small {
    padding: $spacing-xs $spacing-base;
    font-size: $font-size-sm;
  }

  &.ds-button-large {
    padding: $spacing-base $spacing-xl;
    font-size: $font-size-lg;
    height: 88rpx;
  }

  &.ds-loading, &.ds-disabled {
    opacity: 0.6;
    pointer-events: none;
  }

  .ds-loading-icon {
    width: 32rpx;
    height: 32rpx;

    .ds-loading-spinner {
      width: 100%;
      height: 100%;
      border: 4rpx solid rgba(255, 255, 255, 0.3);
      border-top-color: #fff;
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }
  }

  .ds-button-text {
    line-height: 1;
    min-width: 0;
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>