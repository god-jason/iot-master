<template>
  <view class="ds-radio-group">
    <view 
      v-for="(option, index) in options" 
      :key="index"
      class="ds-radio-item"
      :class="{ 'ds-active': modelValue === option.value }"
      @click="handleSelect(option)"
    >
      <view class="ds-radio-circle">
        <view class="ds-radio-inner"></view>
      </view>
      <text class="ds-radio-label">{{ option.label }}</text>
      <DsIcon v-if="option.icon" :name="option.icon" :size="28" color="#999" />
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

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: ''
  },
  options: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const handleSelect = (option) => {
  emit('update:modelValue', option.value)
  emit('change', option)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-base;

  .ds-radio-item {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
    padding: $spacing-sm $spacing-base;
    border-radius: $radius-base;
    background: $bg-secondary;
    transition: $transition-base;

    .ds-radio-circle {
      width: 36rpx;
      height: 36rpx;
      border-radius: 50%;
      border: 4rpx solid #ddd;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: $transition-base;

      .ds-radio-inner {
        width: 20rpx;
        height: 20rpx;
        border-radius: 50%;
        background: $primary-color;
        opacity: 0;
        transform: scale(0);
        transition: $transition-base;
      }
    }

    .ds-radio-label {
      font-size: $font-size-base;
      color: $text-secondary;
    }

    &.ds-active {
      background: $primary-light;

      .ds-radio-circle {
        border-color: $primary-color;

        .ds-radio-inner {
          opacity: 1;
          transform: scale(1);
        }
      }

      .ds-radio-label {
        color: $primary-color;
      }
    }
  }
}
</style>