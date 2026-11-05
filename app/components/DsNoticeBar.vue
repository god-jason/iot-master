<template>
  <view class="ds-notice-bar" :class="[`ds-notice-${type}`]" @click="handleClick">
    <DsIcon v-if="icon" :name="icon" :size="24" class="ds-notice-icon" />
    <scroll-view v-if="scrollable" scroll-x class="ds-notice-scroll">
      <text class="ds-notice-text">{{ text }}</text>
    </scroll-view>
    <text v-else class="ds-notice-text">{{ text }}</text>
    <view v-if="closable" class="ds-notice-close" @click.stop="handleClose">
      <DsIcon name="close" :size="20" />
    </view>
  </view>
</template>

<script setup>
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
    validator: (val) => ['default', 'success', 'warning', 'danger', 'info'].includes(val)
  },
  icon: {
    type: String,
    default: ''
  },
  scrollable: {
    type: Boolean,
    default: true
  },
  closable: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click', 'close'])

const handleClick = (event) => {
  emit('click', event)
}

const handleClose = () => {
  emit('close')
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-notice-bar {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-base;
  background: $bg-secondary;

  &.ds-notice-default {
    background: $bg-secondary;
    color: $text-secondary;
  }

  &.ds-notice-success {
    background: rgba(103, 194, 58, 0.1);
    color: $success-color;
  }

  &.ds-notice-warning {
    background: rgba(230, 162, 60, 0.1);
    color: $warning-color;
  }

  &.ds-notice-danger {
    background: rgba(245, 108, 108, 0.1);
    color: $danger-color;
  }

  &.ds-notice-info {
    background: rgba(91, 163, 232, 0.1);
    color: $primary-color;
  }

  .ds-notice-icon {
    flex-shrink: 0;
  }

  .ds-notice-scroll {
    flex: 1;
    white-space: nowrap;

    .ds-notice-text {
      font-size: $font-size-sm;
      color: inherit;
    }
  }

  .ds-notice-text {
    flex: 1;
    font-size: $font-size-sm;
    color: inherit;
  }

  .ds-notice-close {
    flex-shrink: 0;
    padding: $spacing-xs;
  }
}
</style>