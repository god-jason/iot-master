<template>
  <view v-if="visible" class="ds-dialog-mask" @click="handleMaskClick">
    <view class="ds-dialog-container" @click.stop>
      <view class="ds-dialog-header">
        <text class="ds-dialog-title">{{ title }}</text>
        <view v-if="showClose" class="ds-dialog-close" @click="handleClose">
          <DsIcon name="close" size="32" color="#999" />
        </view>
      </view>
      <view class="ds-dialog-body">
        <slot>{{ message }}</slot>
      </view>
      <view class="ds-dialog-footer">
        <view v-if="showCancel" class="ds-dialog-btn ds-dialog-btn-cancel" @click="handleCancel">
          {{ cancelText }}
        </view>
        <view class="ds-dialog-btn ds-dialog-btn-confirm" @click="handleConfirm">
          {{ confirmText }}
        </view>
      </view>
    </view>
  </view>
</template>

<script setup>
import { watch } from 'vue'
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
  message: {
    type: String,
    default: ''
  },
  confirmText: {
    type: String,
    default: '确定'
  },
  cancelText: {
    type: String,
    default: '取消'
  },
  showCancel: {
    type: Boolean,
    default: true
  },
  showClose: {
    type: Boolean,
    default: false
  },
  closeOnClickMask: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['update:visible', 'confirm', 'cancel', 'close'])

watch(() => props.visible, (val) => {
  if (val) {
    uni.hideKeyboard()
  }
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

const handleConfirm = () => {
  emit('confirm')
  emit('update:visible', false)
}

const handleCancel = () => {
  emit('cancel')
  emit('update:visible', false)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;

  .ds-dialog-container {
    width: 560rpx;
    background: $bg-primary;
    border-radius: $radius-xl;
    overflow: hidden;

    .ds-dialog-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: $spacing-xl;
      border-bottom: 1rpx solid $border-light;

      .ds-dialog-title {
        font-size: $font-size-lg;
        font-weight: 600;
        color: $text-primary;
      }

      .ds-dialog-close {
        padding: $spacing-xs;
      }
    }

    .ds-dialog-body {
      padding: $spacing-xl;
      font-size: $font-size-base;
      color: $text-secondary;
      line-height: 1.6;
    }

    .ds-dialog-footer {
      display: flex;
      border-top: 1rpx solid $border-light;

      .ds-dialog-btn {
        flex: 1;
        padding: $spacing-lg;
        text-align: center;
        font-size: $font-size-base;
        font-weight: 500;

        &.ds-dialog-btn-cancel {
          color: $text-secondary;
          border-right: 1rpx solid $border-light;
        }

        &.ds-dialog-btn-confirm {
          color: $primary-color;
        }

        &:active {
          background: $bg-secondary;
        }
      }
    }
  }
}
</style>