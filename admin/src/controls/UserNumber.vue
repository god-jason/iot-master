<template>
  <div 
    class="control-number" 
    :style="cardStyle"
  >
    <div class="number-header">
      <span class="number-label">{{ control.name }}</span>
      <Icon 
        v-if="iconStyle.icon" 
        :name="iconStyle.icon" 
        :color="iconStyle.color"
        :size="iconStyle.font_size"
        class="number-icon" 
      />
    </div>
    <div class="number-value-wrap">
      <span class="number-value" :style="valueStyle">{{ displayValue }}</span>
      <span class="number-unit" :style="unitStyle">{{ control.unit || '' }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toNum } from './Control.js'
import Icon from './Icon.vue'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number, Boolean],
    default: 0
  }
})

const currentValue = ref(toNum(props.value, 0))

const iconStyle = computed(() => {
  const style = props.control.styles?.icon || {}
  return {
    icon: style.icon || '',
    color: style.color || '#999',
    font_size: style.font_size || '20px'
  }
})

const displayValue = computed(() => {
  const precision = props.control.precision != null ? props.control.precision : 0
  return Number(currentValue.value).toFixed(precision)
})

const cardStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    backgroundColor: style.background_color || '#f5f5f5',
    borderRadius: style.border_radius || '12px'
  }
})

const valueStyle = computed(() => {
  const style = props.control.styles?.text || {}
  return {
    color: style.color || '#333',
    fontSize: style.font_size || '28px'
  }
})

const unitStyle = computed(() => {
  const style = props.control.styles?.text || {}
  return {
    color: style.color ? style.color + '99' : '#999',
    fontSize: style.font_size ? `calc(${style.font_size} * 0.5)` : '14px'
  }
})

watch(() => props.value, (val) => {
  currentValue.value = toNum(val, 0)
})
</script>

<style lang="scss" scoped>
.control-number {
  padding: 16px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100%;
  box-sizing: border-box;

  .number-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .number-label {
      font-size: 13px;
      color: #999;
    }

    .number-icon {
      font-size: 20px;
    }
  }

  .number-value-wrap {
    display: flex;
    align-items: baseline;
    margin-top: 8px;

    .number-value {
      font-weight: 600;
      line-height: 1.2;
    }

    .number-unit {
      margin-left: 4px;
      font-weight: 500;
    }
  }
}
</style>
