<template>
  <view class="ds-section" :class="[`ds-section-${type}`]">
    <view v-if="title" class="ds-section-header">
      <view class="ds-section-title-wrap">
        <view v-if="showLine" class="ds-section-line"></view>
        <view v-if="$slots.icon" class="ds-section-icon">
          <slot name="icon"></slot>
        </view>
        <text class="ds-section-title">{{ title }}</text>
      </view>
      <view class="ds-section-extra">
        <slot name="extra"></slot>
      </view>
    </view>
    <view class="ds-section-body">
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
  },
  type: {
    type: String,
    default: 'default',
    validator: (val) => ['default', 'card', 'plain'].includes(val)
  },
  showLine: {
    type: Boolean,
    default: true
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-section {
  margin: $spacing-base;
  padding: $spacing-base;

  &.ds-section-card {
    background: $bg-primary;
    border-radius: $radius-lg;
    padding: $spacing-lg;
    box-shadow: $shadow-sm;
  }

  &.ds-section-plain {
    margin-bottom: $spacing-base;
  }

  .ds-section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: $spacing-base;

      .ds-section-title-wrap {
        display: flex;
        align-items: center;
        gap: $spacing-xs;

        .ds-section-icon {
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .ds-section-line {
        width: 8rpx;
        height: 32rpx;
        background: $primary-color;
        border-radius: 4rpx;
      }

      .ds-section-title {
        font-size: $font-size-lg;
        font-weight: 600;
        color: $text-primary;
      }
    }

    .ds-section-extra {
      font-size: $font-size-sm;
      color: $text-secondary;
    }
  }

  .ds-section-body {
    padding: $spacing-base;

    .ds-section-card & {
      padding-top: 0;
    }
  }
}
</style>