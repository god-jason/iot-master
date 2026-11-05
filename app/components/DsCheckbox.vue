<template>
  <view class="ds-checkbox-group">
    <view 
      v-for="(option, index) in options" 
      :key="index"
      class="ds-checkbox-item"
      :class="{ 'ds-active': isSelected(option.value) }"
      @click="handleToggle(option)"
    >
      <view class="ds-checkbox-box">
        <DsIcon v-if="isSelected(option.value)" name="affirm" :size="24" color="#fff" />
      </view>
      <text class="ds-checkbox-label">{{ option.label }}</text>
      <DsIcon v-if="option.icon" :name="option.icon" :size="28" color="#999" />
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
  modelValue: {
    type: Array,
    default: () => []
  },
  options: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const isSelected = (value) => {
  return props.modelValue.includes(value)
}

const handleToggle = (option) => {
  const newValue = [...props.modelValue]
  const index = newValue.indexOf(option.value)
  if (index > -1) {
    newValue.splice(index, 1)
  } else {
    newValue.push(option.value)
  }
  emit('update:modelValue', newValue)
  emit('change', newValue)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-base;

  .ds-checkbox-item {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
    padding: $spacing-sm $spacing-base;
    border-radius: $radius-base;
    background: $bg-secondary;
    transition: $transition-base;

    .ds-checkbox-box {
      width: 36rpx;
      height: 36rpx;
      border-radius: $radius-sm;
      border: 4rpx solid #ddd;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #fff;
      transition: $transition-base;
    }

    .ds-checkbox-label {
      font-size: $font-size-base;
      color: $text-secondary;
    }

    &.ds-active {
      background: $primary-light;

      .ds-checkbox-box {
        background: $primary-color;
        border-color: $primary-color;
      }

      .ds-checkbox-label {
        color: $primary-color;
      }
    }
  }
}
</style>