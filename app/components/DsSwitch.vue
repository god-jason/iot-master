<template>
  <view 
    class="ds-switch" 
    :class="{ 'ds-active': currentValue }"
    :style="switchStyle"
    @click="handleToggle"
  >
    <view class="ds-switch-track">
      <view class="ds-switch-thumb"></view>
    </view>
    <text v-if="label" class="ds-switch-label">{{ label }}</text>
  </view>
</template>

<script setup>
import { computed } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: undefined
  },
  value: {
    type: Boolean,
    default: false
  },
  label: {
    type: String,
    default: ''
  },
  activeColor: {
    type: String,
    default: '#5BA3E8'
  },
  inactiveColor: {
    type: String,
    default: '#e0e0e0'
  }
})

const emit = defineEmits(['change', 'update:modelValue'])

const currentValue = computed(() => (props.modelValue !== undefined ? props.modelValue : props.value))

const switchStyle = computed(() => ({
  '--switch-active-color': props.activeColor,
  '--switch-inactive-color': props.inactiveColor
}))

const handleToggle = () => {
  const nextValue = !currentValue.value
  emit('update:modelValue', nextValue)
  emit('change', nextValue)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-switch {
  display: inline-flex;
  align-items: center;
  gap: $spacing-sm;

  .ds-switch-track {
    width: 120rpx;
    height: 64rpx;
    background: var(--switch-inactive-color, #e0e0e0);
    border-radius: 32rpx;
    position: relative;
    transition: $transition-base;

    .ds-switch-thumb {
      width: 56rpx;
      height: 56rpx;
      background: #fff;
      border-radius: 50%;
      position: absolute;
      top: 4rpx;
      left: 4rpx;
      box-shadow: 0 2rpx 6rpx rgba(0, 0, 0, 0.15);
      transition: $transition-base;
    }
  }

  .ds-switch-label {
    font-size: $font-size-base;
    color: $text-primary;
  }

  &.ds-active {
    .ds-switch-track {
      background: var(--switch-active-color, #5BA3E8);

      .ds-switch-thumb {
        transform: translateX(56rpx);
      }
    }
  }
}
</style>