<template>
  <view class="ds-description">
    <view v-if="title" class="ds-description-header">
      <text class="ds-description-title">{{ title }}</text>
    </view>
    <view class="ds-description-body">
      <slot></slot>
    </view>
  </view>
</template>

<script setup>
defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps({
  title: {
    type: String,
    default: ''
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-description {
  .ds-description-header {
    padding-bottom: $spacing-sm;
    margin-bottom: $spacing-base;
    border-bottom: 1rpx solid $border-light;

    .ds-description-title {
      font-size: $font-size-base;
      font-weight: 600;
      color: $text-primary;
    }
  }

  .ds-description-body {
    .ds-description-item {
      display: flex;
      padding: $spacing-sm 0;
    }
  }
}
</style>