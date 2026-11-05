<template>
  <view class="ds-navbar" :class="{ 'ds-transparent': transparent, 'ds-status-bar': statusBar }" @click="handleClick">
    <view class="ds-navbar-status" v-if="statusBar"></view>
    <view class="ds-navbar-content">
      <view v-if="showBack" class="ds-navbar-back" @click.stop="handleBack">
        <DsIcon name="back" :size="24" :color="color" />
      </view>
      <view class="ds-navbar-left">
        <slot name="left"></slot>
	  <!-- #ifdef MP -->
	    <slot name="right"></slot>
	  <!-- #endif -->
	  </view>
      <view class="ds-navbar-title">
        <text>{{ title }}</text>
      </view>
      <view class="ds-navbar-right">
      <!-- #ifndef MP -->
        <slot name="right"></slot>
      <!-- #endif -->
      </view>
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

defineProps({
  title: {
    type: String,
    default: ''
  },
  showBack: {
    type: Boolean,
    default: true
  },
  transparent: {
    type: Boolean,
    default: false
  },
  statusBar: {
    type: Boolean,
    default: true
  },
  color: {
    type: String,
    default: '#333'
  }
})

const emit = defineEmits(['click'])

const handleClick = (event) => {
  emit('click', event)
}

const handleBack = () => {
  uni.navigateBack({
    fail: () => {
      uni.reLaunch({ url: '/pages/device/index' })
    }
  })
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-navbar {
  background: $bg-primary;
  position: sticky;
  top: 0;
  left: 0;
  right: 0;
  
  z-index: 9999;

  &.ds-transparent {
    background: transparent;
  }

  &.ds-status-bar {
    padding-top: var(--status-bar-height);
  }

  .ds-navbar-status {
    height: var(--status-bar-height);
    width: 100%;
    background: inherit;
  }

  .ds-navbar-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 88rpx;
    padding: 0 $spacing-base;

    .ds-navbar-back {
      width: 60rpx;
      display: flex;
      align-items: center;
    }
	
    .ds-navbar-left {
      width: 120rpx;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .ds-navbar-title {
      flex: 1;
      text-align: center;

      text {
        font-size: $font-size-lg;
        font-weight: 600;
        color: $text-primary;
      }
    }

    .ds-navbar-right {
      width: 120rpx;
      display: flex;
      align-items: center;
      justify-content: flex-end;
    }
  }
}
</style>