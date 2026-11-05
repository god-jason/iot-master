<template>
  <view class="ds-result" @click="handleClick">
    <view class="ds-result-icon" :class="[`ds-result-${type}`]">
      <DsIcon :name="icon" :size="80" />
    </view>
    <text class="ds-result-title">{{ title }}</text>
    <text v-if="message" class="ds-result-message">{{ message }}</text>
    <view v-if="actions.length" class="ds-result-actions">
      <DsButton 
        v-for="(action, index) in actions" 
        :key="index"
        :text="action.text"
        :type="action.type || 'primary'"
        :size="action.size || 'medium'"
        @click.stop="action.onClick"
      />
    </view>
    <view v-if="showHome" class="ds-result-home">
      <text @click.stop="handleHome">返回首页</text>
    </view>
  </view>
</template>

<script setup>
import DsIcon from './DsIcon.vue'
import DsButton from './DsButton.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps({
  type: {
    type: String,
    default: 'success',
    validator: (val) => ['success', 'fail', 'warning', 'info', 'empty'].includes(val)
  },
  title: {
    type: String,
    default: ''
  },
  message: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: ''
  },
  actions: {
    type: Array,
    default: () => []
  },
  showHome: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

const handleClick = (event) => {
  emit('click', event)
}

const handleHome = () => {
  uni.switchTab({
    url: '/pages/device/index',
    fail: () => {
      uni.reLaunch({ url: '/pages/device/index' })
    }
  })
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-result {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: $spacing-xl * 2;

  .ds-result-icon {
    width: 160rpx;
    height: 160rpx;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: $spacing-lg;

    &.ds-result-success {
      background: rgba(103, 194, 58, 0.1);
      color: $success-color;
    }

    &.ds-result-fail {
      background: rgba(245, 108, 108, 0.1);
      color: $danger-color;
    }

    &.ds-result-warning {
      background: rgba(230, 162, 60, 0.1);
      color: $warning-color;
    }

    &.ds-result-info {
      background: rgba(91, 163, 232, 0.1);
      color: $primary-color;
    }

    &.ds-result-empty {
      background: rgba(153, 153, 153, 0.1);
      color: $text-placeholder;
    }
  }

  .ds-result-title {
    font-size: $font-size-xl;
    font-weight: 600;
    color: $text-primary;
    margin-bottom: $spacing-sm;
  }

  .ds-result-message {
    font-size: $font-size-base;
    color: $text-secondary;
    margin-bottom: $spacing-xl;
    text-align: center;
  }

  .ds-result-actions {
    display: flex;
    gap: $spacing-base;
    margin-bottom: $spacing-xl;
  }

  .ds-result-home {
    text {
      font-size: $font-size-sm;
      color: $text-placeholder;

      &:active {
        color: $primary-color;
      }
    }
  }
}
</style>