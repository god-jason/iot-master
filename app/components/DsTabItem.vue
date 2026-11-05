<template>
  <view 
    class="ds-tab-item"
    :class="{ 'ds-active': isActive }"
    @click="handleClick"
  >
    <text class="ds-tab-text">{{ label }}</text>
    <view v-if="badge" class="ds-tab-badge">{{ badge }}</view>
    <view v-if="isActive && $parent.$props.type === 'line'" class="ds-tab-indicator"></view>
    <slot></slot>
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
  label: {
    type: String,
    default: ''
  },
  badge: {
    type: [String, Number],
    default: null
  },
  index: {
    type: Number,
    default: 0
  },
  active: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

const isActive = computed(() => {
  return props.active
})

const handleClick = () => {
  emit('click', props.index)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-tab-item {
  display: inline-flex;
  align-items: center;
  padding: $spacing-sm $spacing-lg;
  position: relative;
  transition: $transition-base;

  .ds-tab-text {
    font-size: $font-size-base;
    color: $text-secondary;
    transition: $transition-base;
  }

  .ds-tab-badge {
    min-width: 28rpx;
    height: 28rpx;
    padding: 0 8rpx;
    background: $danger-color;
    color: #fff;
    font-size: $font-size-xs;
    border-radius: 14rpx;
    margin-left: $spacing-xs;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  &.ds-active {
    .ds-tab-text {
      color: $primary-color;
      font-weight: 600;
    }
  }

  &:active {
    opacity: 0.7;
  }
}

.ds-tab-indicator {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 6rpx;
  background: $primary-color;
  border-radius: 3rpx;
}
</style>