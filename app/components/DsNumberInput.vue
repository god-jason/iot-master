<template>
  <view class="ds-number-input">
    <view class="ds-number-btn" :class="{ 'ds-disabled': isMin }" @click="handleDecrease">
      <DsIcon name="reduce" size="24" color="#5BA3E8" />
    </view>
    <input
      :value="displayValue"
      type="digit"
      :disabled="disabled"
      class="ds-number-field"
      @input="handleInput"
      @blur="handleBlur"
    />
    <view class="ds-number-btn" :class="{ 'ds-disabled': isMax }" @click="handleIncrease">
      <DsIcon name="add" size="24" color="#5BA3E8" />
    </view>
    <text v-if="unit" class="ds-number-unit">{{ unit }}</text>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import DsIcon from './DsIcon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  modelValue: {
    type: [Number, String],
    default: 0
  },
  min: {
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    default: 99999
  },
  step: {
    type: Number,
    default: 1
  },
  precision: {
    type: Number,
    default: 0
  },
  unit: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const currentValue = ref(Number(props.modelValue) || 0)

const displayValue = computed(() => {
  return Number(currentValue.value).toFixed(props.precision)
})

const isMin = computed(() => currentValue.value <= props.min)
const isMax = computed(() => currentValue.value >= props.max)

watch(() => props.modelValue, (val) => {
  currentValue.value = Number(val) || 0
})

const handleInput = (e) => {
  let val = parseFloat(e.detail.value) || 0
  if (val < props.min) val = props.min
  if (val > props.max) val = props.max
  currentValue.value = val
  emit('update:modelValue', currentValue.value)
}

const handleBlur = () => {
  emit('change', currentValue.value)
}

const handleDecrease = () => {
  if (isMin.value || props.disabled) return
  currentValue.value = Math.max(props.min, currentValue.value - props.step)
  emit('update:modelValue', currentValue.value)
  emit('change', currentValue.value)
}

const handleIncrease = () => {
  if (isMax.value || props.disabled) return
  currentValue.value = Math.min(props.max, currentValue.value + props.step)
  emit('update:modelValue', currentValue.value)
  emit('change', currentValue.value)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-number-input {
  display: flex;
  align-items: center;
  gap: $spacing-sm;

  .ds-number-btn {
    width: 64rpx;
    height: 64rpx;
    border-radius: 50%;
    background: $bg-secondary;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: $transition-base;

    &.ds-disabled {
      opacity: 0.4;
      pointer-events: none;
    }

    &:active {
      background: $primary-light;
    }
  }

  .ds-number-field {
    width: 160rpx;
    text-align: center;
    font-size: $font-size-lg;
    color: $text-primary;
    background: transparent;
  }

  .ds-number-unit {
    font-size: $font-size-sm;
    color: $text-placeholder;
    margin-left: $spacing-xs;
  }
}
</style>