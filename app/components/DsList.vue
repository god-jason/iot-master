<template>
  <view class="ds-list" :class="{ 'ds-bordered': bordered, 'ds-striped': striped }">
    <slot></slot>
  </view>
</template>

<script setup>
defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps({
  bordered: {
    type: Boolean,
    default: false
  },
  striped: {
    type: Boolean,
    default: false
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-list {
  background: $bg-primary;
  border-radius: $radius-lg;
  overflow: hidden;

  &.ds-bordered {
    border: 2rpx solid $border-color;
  }

  &.ds-striped {
    .ds-list-item {
      &:nth-child(even) {
        background: $bg-secondary;
      }
    }
  }
}
</style>