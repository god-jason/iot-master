<template>
  <view class="ds-loading" :class="[`ds-loading-${type}`]">
    <view v-if="type === 'spinner'" class="ds-loading-spinner">
      <view class="ds-spinner-ring"></view>
    </view>
    <view v-else-if="type === 'dots'" class="ds-loading-dots">
      <view v-for="i in 3" :key="i" class="ds-dot" :style="{ animationDelay: `${(i - 1) * 0.2}s` }"></view>
    </view>
    <view v-else class="ds-loading-ring">
      <view class="ds-ring"></view>
    </view>
    <text v-if="text" class="ds-loading-text">{{ text }}</text>
  </view>
</template>

<script setup>
defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps({
  text: {
    type: String,
    default: '加载中...'
  },
  type: {
    type: String,
    default: 'ring',
    validator: (val) => ['ring', 'spinner', 'dots'].includes(val)
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: $spacing-xl;

  .ds-loading-spinner {
    width: 64rpx;
    height: 64rpx;

    .ds-spinner-ring {
      width: 100%;
      height: 100%;
      border: 6rpx solid rgba(91, 163, 232, 0.2);
      border-top-color: $primary-color;
      border-radius: 50%;
      animation: spin 1s linear infinite;
    }
  }

  .ds-loading-dots {
    display: flex;
    gap: 16rpx;

    .ds-dot {
      width: 16rpx;
      height: 16rpx;
      background: $primary-color;
      border-radius: 50%;
      animation: bounce 0.6s ease-in-out infinite;
    }
  }

  .ds-loading-ring {
    width: 64rpx;
    height: 64rpx;

    .ds-ring {
      width: 100%;
      height: 100%;
      border: 4rpx solid transparent;
      border-top-color: $primary-color;
      border-right-color: $primary-color;
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }
  }

  .ds-loading-text {
    margin-top: $spacing-base;
    font-size: $font-size-sm;
    color: $text-placeholder;
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes bounce {
  0%, 100% {
    transform: scale(1);
    opacity: 0.6;
  }
  50% {
    transform: scale(1.5);
    opacity: 1;
  }
}
</style>