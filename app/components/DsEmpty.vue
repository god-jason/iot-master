<template>
  <view class="ds-empty" :class="[`ds-empty-${type}`]" @click="handleClick">
    <view class="ds-empty-icon">
      <DsIcon :name="icon" :size="iconSize" :color="iconColor" />
    </view>
    <text class="ds-empty-text">{{ text }}</text>
    <text v-if="subText" class="ds-empty-subtext">{{ subText }}</text>
    <view v-if="$slots.action" class="ds-empty-action">
      <slot name="action"></slot>
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
  text: {
    type: String,
    default: '暂无数据'
  },
  subText: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: ''
  },
  iconSize: {
    type: Number,
    default: 80
  },
  iconColor: {
    type: String,
    default: '#ccc'
  },
  type: {
    type: String,
    default: 'default',
    validator: (val) => ['default', 'search', 'error', 'network'].includes(val)
  }
})

const emit = defineEmits(['click'])

const icon = computed(() => {
  const icons = {
    default: 'store',
    search: 'search',
    error: 'warning',
    network: 'wifi'
  }
  return props.icon || icons[props.type]
})

const handleClick = (event) => {
  emit('click', event)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: $spacing-xl * 2;

  .ds-empty-icon {
    width: 160rpx;
    height: 160rpx;
    border-radius: 50%;
    background: $bg-secondary;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: $spacing-lg;
  }

  .ds-empty-text {
    font-size: $font-size-lg;
    color: $text-placeholder;
    margin-bottom: $spacing-xs;
  }

  .ds-empty-subtext {
    font-size: $font-size-sm;
    color: $text-disabled;
    margin-bottom: $spacing-lg;
  }

  .ds-empty-action {
    margin-top: $spacing-base;
  }
}
</style>