<template>
  <view 
    class="ds-avatar-wrapper" 
    :class="[`ds-avatar-${shape}`, `ds-avatar-${size}`]"
    @click="handleClick"
  >
    <view class="ds-avatar-inner" :style="avatarStyle">
      <image v-if="src" :src="src" mode="aspectFill" class="ds-avatar-image" />
      <DsIcon v-else :name="icon" :size="iconSize" :color="iconColor" />
    </view>
    <view v-if="badge" class="ds-avatar-badge" :class="`ds-badge-${badge.type || 'dot'}`">
      <text v-if="badge.value" class="ds-badge-text">{{ badge.value }}</text>
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
  src: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: 'my'
  },
  shape: {
    type: String,
    default: 'circle',
    validator: (val) => ['circle', 'square'].includes(val)
  },
  size: {
    type: String,
    default: 'medium',
    validator: (val) => ['mini', 'small', 'medium', 'large', 'xlarge'].includes(val)
  },
  iconColor: {
    type: String,
    default: '#fff'
  },
  bgColor: {
    type: String,
    default: '#5BA3E8'
  },
  badge: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['click'])

const iconSize = computed(() => {
  const sizeMap = {
    mini: 24,
    small: 32,
    medium: 40,
    large: 48,
    xlarge: 64
  }
  return sizeMap[props.size] || 40
})

const avatarStyle = computed(() => ({
  background: props.src ? 'transparent' : props.bgColor
}))

const handleClick = (e) => {
  emit('click', e)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-avatar-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;

  &.ds-avatar-circle {
    .ds-avatar-inner {
      border-radius: 50%;
    }
  }

  &.ds-avatar-square {
    .ds-avatar-inner {
      border-radius: $radius-sm;
    }
  }

  &.ds-avatar-mini {
    .ds-avatar-inner {
      width: 64rpx;
      height: 64rpx;
    }
  }

  &.ds-avatar-small {
    .ds-avatar-inner {
      width: 80rpx;
      height: 80rpx;
    }
  }

  &.ds-avatar-medium {
    .ds-avatar-inner {
      width: 100rpx;
      height: 100rpx;
    }
  }

  &.ds-avatar-large {
    .ds-avatar-inner {
      width: 120rpx;
      height: 120rpx;
    }
  }

  &.ds-avatar-xlarge {
    .ds-avatar-inner {
      width: 160rpx;
      height: 160rpx;
    }
  }
}

.ds-avatar-inner {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;

  .ds-avatar-image {
    width: 100%;
    height: 100%;
  }
}

.ds-avatar-badge {
  position: absolute;
  top: -4rpx;
  right: -4rpx;
  min-width: 32rpx;
  height: 32rpx;
  background: $danger-color;
  border-radius: 16rpx;
  border: 4rpx solid $bg-primary;
  display: flex;
  align-items: center;
  justify-content: center;

  &.ds-badge-dot {
    width: 24rpx;
    height: 24rpx;
    border-radius: 50%;
    min-width: auto;
  }

  .ds-badge-text {
    font-size: $font-size-xs;
    color: #fff;
    padding: 0 8rpx;
  }
}
</style>