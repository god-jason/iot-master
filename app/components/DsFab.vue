<template>
  <view 
    class="ds-fab" 
    :class="[`ds-fab-${position}`, { 'ds-show': show, 'ds-mini': mini, 'ds-has-text': !!text }]"
    @click="handleClick"
  >
    <DsIcon :name="icon" :size="iconSize" color="#fff" />
    <text v-if="text" class="ds-fab-text">{{ text }}</text>
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
  icon: {
    type: String,
    default: 'plus'
  },
  text: {
    type: String,
    default: ''
  },
  position: {
    type: String,
    default: 'right-bottom',
    validator: (val) => ['right-bottom', 'left-bottom', 'right-top', 'left-top'].includes(val)
  },
  mini: {
    type: Boolean,
    default: false
  },
  show: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['click'])

const iconSize = computed(() => props.mini ? 28 : 36)

const handleClick = () => {
  emit('click')
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-fab {
  position: fixed;
  width: 100rpx;
  height: 100rpx;
  background: $primary-color;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: $shadow-lg;
  z-index: 100;
  transition: $transition-base;

  &.ds-fab-right-bottom {
    right: $spacing-lg;
    bottom: calc(#{$spacing-xl} + #{$safe-bottom});
  }

  &.ds-fab-left-bottom {
    left: $spacing-lg;
    bottom: calc(#{$spacing-xl} + #{$safe-bottom});
  }

  &.ds-fab-right-top {
    right: $spacing-lg;
    top: calc(#{$spacing-xl} + var(--status-bar-height));
  }

  &.ds-fab-left-top {
    left: $spacing-lg;
    top: calc(#{$spacing-xl} + var(--status-bar-height));
  }

  &.ds-mini {
    width: 72rpx;
    height: 72rpx;
  }

  &.ds-has-text {
    width: auto;
    min-width: 100rpx;
    padding: 0 24rpx;
    gap: 12rpx;
    border-radius: 999rpx;

    &.ds-mini {
      min-width: 72rpx;
      padding: 0 20rpx;
    }
  }

  &:not(.ds-show) {
    opacity: 0;
    transform: translateY(20rpx);
    pointer-events: none;
  }

  &:active {
    transform: scale(0.95);
  }

  .ds-fab-text {
    font-size: $font-size-sm;
    color: #fff;
    white-space: nowrap;
  }
}
</style>