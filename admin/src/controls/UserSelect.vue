<template>
  <div class="control-mode-select">
    <span class="mode-label">{{ control.name }}</span>
    <div class="mode-options">
      <div 
        v-for="option in control.options || defaultOptions" 
        :key="option.value"
        class="mode-option"
        :class="{ active: currentValue === option.value }"
        :style="getOptionStyle(currentValue === option.value)"
        @click="select(option.value)"
      >
        <Icon :name="option.icon" class="option-icon" />
        <span class="option-label">{{ option.label }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { toStr } from './Control.js'
import Icon from './Icon.vue'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number, Boolean],
    default: 'auto'
  }
})

const emit = defineEmits(['change', 'action'])

const defaultOptions = [
  { label: '自动', value: 'auto', icon: 'scene' },
      { label: '手动', value: 'manual', icon: 'illuminance' }
]

const currentValue = ref(toStr(props.value))

watch(() => props.value, (val) => {
  currentValue.value = toStr(val)
})

const getOptionStyle = (isActive) => {
  const activeStyle = props.control.styles?.active || {}
  const inactiveStyle = props.control.styles?.inactive || {}
  return isActive ? {
    backgroundColor: activeStyle.background_color || '#5BA3E8',
    color: activeStyle.color || '#ffffff',
    borderRadius: activeStyle.border_radius || '6px'
  } : {
    color: inactiveStyle.color || '#666666'
  }
}

const select = (value) => {
  currentValue.value = value
  emit('change', { type: 'mode', value: value, control: props.control })
  if (props.control.events?.change?.key) {
    emit('action', { 
      type: 'change', 
      control: props.control,
      eventKey: 'change',
    })
  }
}
</script>

<style lang="scss" scoped>
.control-mode-select {
  .mode-label {
    font-size: 14px;
    color: #333;
    display: block;
    margin-bottom: 12px;
  }

  .mode-options {
    display: flex;
    gap: 8px;

    .mode-option {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 12px;
      background: #f5f5f5;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.3s;

      .option-icon {
        font-size: 24px;
        margin-bottom: 4px;
      }

      .option-label {
        font-size: 12px;
      }

      &.active {
        background: #5BA3E8;
        color: #fff;
      }
    }
  }
}
</style>
