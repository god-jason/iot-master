<template>
  <view v-if="visible" class="ds-drawer-mask" @click="handleMaskClick">
    <view class="ds-drawer-container" :class="[`ds-drawer-${direction}`]" :style="drawerStyle" @click.stop>
      <view v-if="title" class="ds-drawer-header">
        <text class="ds-drawer-title">{{ title }}</text>
        <view class="ds-drawer-close" @click.stop="handleClose">
          <DsIcon name="close" size="32" color="#999" />
        </view>
      </view>
      <scroll-view scroll-y class="ds-drawer-body">
        <slot></slot>
      </scroll-view>
      <view v-if="$slots.footer" class="ds-drawer-footer">
        <slot name="footer"></slot>
      </view>
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
  visible: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: ''
  },
  direction: {
    type: String,
    default: 'right',
    validator: (val) => ['left', 'right', 'top', 'bottom'].includes(val)
  },
  width: {
    type: [String, Number],
    default: '80%'
  },
  height: {
    type: [String, Number],
    default: '80%'
  },
  closeOnClickMask: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['update:visible', 'close'])

const drawerStyle = computed(() => {
  const style = {}
  if (props.direction === 'left' || props.direction === 'right') {
    style.width = typeof props.width === 'number' ? `${props.width}rpx` : props.width
  } else {
    style.height = typeof props.height === 'number' ? `${props.height}rpx` : props.height
  }
  return style
})

const handleMaskClick = () => {
  if (props.closeOnClickMask) {
    handleClose()
  }
}

const handleClose = () => {
  emit('update:visible', false)
  emit('close')
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-drawer-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;

  .ds-drawer-container {
    position: absolute;
    background: $bg-primary;
    height: 100%;

    &.ds-drawer-left {
      left: 0;
      top: 0;
      border-radius: 0 $radius-xl $radius-xl 0;
    }

    &.ds-drawer-right {
      right: 0;
      top: 0;
      border-radius: $radius-xl 0 0 $radius-xl;
    }

    &.ds-drawer-top {
      top: 0;
      left: 0;
      right: 0;
      height: auto;
      border-radius: 0 0 $radius-xl $radius-xl;
    }

    &.ds-drawer-bottom {
      bottom: 0;
      left: 0;
      right: 0;
      height: auto;
      border-radius: $radius-xl $radius-xl 0 0;
    }

    .ds-drawer-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: $spacing-xl;
      border-bottom: 1rpx solid $border-light;

      .ds-drawer-title {
        font-size: $font-size-lg;
        font-weight: 600;
        color: $text-primary;
      }

      .ds-drawer-close {
        padding: $spacing-xs;
      }
    }

    .ds-drawer-body {
      height: calc(100% - 100rpx);

      .ds-drawer-top &,
      .ds-drawer-bottom & {
        height: 100%;
      }
    }

    .ds-drawer-footer {
      padding: $spacing-xl;
      border-top: 1rpx solid $border-light;
      padding-bottom: calc(#{$spacing-xl} + #{$safe-bottom});
    }
  }
}
</style>