<template>
  <view class="ds-description-item">
    <text class="ds-description-label">{{ label }}</text>
    <view class="ds-description-value">
      <slot>{{ value }}</slot>
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
  label: {
    type: String,
    default: ''
  },
  value: {
    type: [String, Number],
    default: ''
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-description-item {
  display: flex;
  padding: $spacing-sm 0;

  .ds-description-label {
    width: 160rpx;
    font-size: $font-size-sm;
    color: $text-placeholder;
    flex-shrink: 0;
  }

  .ds-description-value {
    flex: 1;
    font-size: $font-size-sm;
    color: $text-primary;
  }
}
</style>