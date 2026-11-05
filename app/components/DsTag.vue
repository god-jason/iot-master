<template>
  <view class="ds-tag" :class="[`ds-tag-${type}`, `ds-tag-${size}`]" @click="handleClick">
    <DsIcon v-if="icon" :name="icon" :size="iconSize" />
    <text><slot>{{ text }}</slot></text>
    <view v-if="closable" class="ds-tag-close" @click.stop="handleClose">
      <DsIcon name="close" :size="16" />
    </view>
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
    default: 'default',
    validator: (val) => ['default', 'primary', 'success', 'warning', 'danger'].includes(val)
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
  closable: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click', 'close'])

const iconSize = computed(() => {
  const sizes = { mini: 14, small: 16, medium: 20, large: 24 }
  return sizes[props.size] || 20
})

const handleClick = (event) => {
  emit('click', event)
}

const handleClose = () => {
  emit('close')
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-tag {
  display: inline-flex;
  align-items: center;
  gap: $spacing-xs;
  padding: $spacing-xs $spacing-sm;
  border-radius: $radius-sm;
  font-size: $font-size-sm;

  &.ds-tag-mini {
    padding: 2rpx 8rpx;
    font-size: 20rpx;
  }

  &.ds-tag-small {
    padding: 4rpx $spacing-xs;
    font-size: $font-size-xs;
  }

  &.ds-tag-large {
    padding: $spacing-sm $spacing-base;
    font-size: $font-size-base;
  }

  &.ds-tag-default {
    background: $bg-secondary;
    color: $text-secondary;
  }

  &.ds-tag-primary {
    background: $primary-light;
    color: $primary-color;
  }

  &.ds-tag-success {
    background: rgba(103, 194, 58, 0.1);
    color: $success-color;
  }

  &.ds-tag-warning {
    background: rgba(230, 162, 60, 0.1);
    color: $warning-color;
  }

  &.ds-tag-danger {
    background: rgba(245, 108, 108, 0.1);
    color: $danger-color;
  }

  .ds-tag-close {
    padding: 2rpx;
  }
}
</style>