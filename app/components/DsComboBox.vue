<template>
  <picker :range="options" range-key="label" :value="selectedIndex" @change="onChange">
    <view class="ds-combo-display">
      <text v-if="selectedLabel" class="ds-combo-text">{{ selectedLabel }}</text>
      <text v-else class="ds-combo-placeholder">{{ placeholder || label }}</text>
      <DsIcon name="next" :size="16" color="#999" />
    </view>
  </picker>
</template>

<script setup>
import { computed } from 'vue'
import DsIcon from './DsIcon.vue'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  value: { type: [String, Number], default: '' },
  options: { type: Array, default: () => [] },
  placeholder: { type: String, default: '请选择' },
  label: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue', 'change'])

const currentValue = computed(() => {
  if (props.modelValue !== '' && props.modelValue !== null && props.modelValue !== undefined) {
    return props.modelValue
  }
  return props.value
})

const selectedIndex = computed(() => {
  const idx = props.options.findIndex(o => o.value === currentValue.value)
  return idx >= 0 ? idx : 0
})

const selectedLabel = computed(() => {
  const item = props.options.find(o => o.value === currentValue.value)
  return item?.label || ''
})

const onChange = (e) => {
  const idx = e.detail.value
  const option = props.options[idx]
  const val = option?.value ?? ''
  emit('update:modelValue', val)
  emit('change', val)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-combo-display {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8rpx;
  min-height: 72rpx;
}

.ds-combo-text {
  font-size: $font-size-base;
  color: $text-primary;
}

.ds-combo-placeholder {
  font-size: $font-size-base;
  color: $text-placeholder;
}
</style>