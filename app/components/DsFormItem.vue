<template>
  <view class="ds-form-item">
    <view v-if="label" class="ds-form-item-label">
      <text>{{ label }}</text>
      <text v-if="required" class="ds-form-item-required">*</text>
    </view>
    <view class="ds-form-item-content">
      <slot></slot>
    </view>
    <text v-if="error" class="ds-form-item-error">{{ error }}</text>
    <text v-if="help" class="ds-form-item-help">{{ help }}</text>
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
  required: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  },
  help: {
    type: String,
    default: ''
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-form-item {
  padding: $spacing-base 0;

  .ds-form-item-label {
    display: flex;
    align-items: center;
    gap: $spacing-xs;
    margin-bottom: $spacing-sm;
    font-size: $font-size-base;
    color: $text-primary;

    .ds-form-item-required {
      color: $danger-color;
      font-size: $font-size-lg;
    }
  }

  .ds-form-item-content {
    width: 100%;
  }

  .ds-form-item-error {
    font-size: $font-size-xs;
    color: $danger-color;
    margin-top: $spacing-xs;
    display: block;
  }

  .ds-form-item-help {
    font-size: $font-size-xs;
    color: $text-placeholder;
    margin-top: $spacing-xs;
    display: block;
  }
}
</style>