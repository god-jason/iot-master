<template>
  <view class="ds-slider">
    <view v-if="showLabel" class="ds-slider-header">
      <text class="ds-slider-label">{{ label }}</text>
      <text class="ds-slider-value">{{ currentValue }}{{ unit }}</text>
    </view>
    <slider
      :min="min"
      :max="max"
      :step="step"
      :value="currentValue"
      :activeColor="color"
      backgroundColor="#e2e8f0"
      block-size="24"
      @change="handleChange"
      @changing="handleChanging"
      class="ds-slider-native"
    />
    <view v-if="showRange" class="ds-slider-range">
      <text>{{ min }}{{ unit }}</text>
      <text>{{ max }}{{ unit }}</text>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  modelValue: {
    type: [Number, String],
    default: 50
  },
  min: {
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    default: 100
  },
  step: {
    type: Number,
    default: 1
  },
  label: {
    type: String,
    default: ''
  },
  unit: {
    type: String,
    default: ''
  },
  color: {
    type: String,
    default: '#5BA3E8'
  },
  showLabel: {
    type: Boolean,
    default: false
  },
  showRange: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const currentValue = ref(Number(props.modelValue))

watch(() => props.modelValue, (val) => {
  currentValue.value = Number(val)
})

const fillPercent = computed(() => {
  return ((currentValue.value - props.min) / (props.max - props.min)) * 100
})

const handleChange = (e) => {
  currentValue.value = e.detail.value
  emit('update:modelValue', currentValue.value)
  emit('change', currentValue.value)
}

const handleChanging = (e) => {
  currentValue.value = e.detail.value
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-slider {
  .ds-slider-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: $spacing-sm;

    .ds-slider-label {
      font-size: $font-size-sm;
      color: $text-secondary;
    }

    .ds-slider-value {
      font-size: $font-size-base;
      font-weight: 600;
      color: $primary-color;
    }
  }

  .ds-slider-native {
    margin: $spacing-sm 0;
  }

  .ds-slider-range {
    display: flex;
    justify-content: space-between;
    font-size: $font-size-xs;
    color: $text-placeholder;
  }
}
</style>