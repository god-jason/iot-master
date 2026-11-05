<template>
  <div 
    class="control-button" 
    :style="buttonStyle"
    @click="handleClick"
  >
    <Icon 
      v-if="iconStyle.icon" 
      :name="iconStyle.icon" 
      :color="iconStyle.color"
      :size="iconStyle.font_size"
      class="button-icon" 
    />
    <span class="button-text">{{ control.name }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import Icon from './Icon.vue'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number],
    default: ''
  }
})

const emit = defineEmits(['change', 'action'])

const buttonStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    backgroundColor: style.background_color || '#5BA3E8',
    color: style.color || '#fff',
    fontSize: style.font_size || '16px',
    borderRadius: style.border_radius || '8px',
    border: style.border || 'none',
    padding: style.padding || '16px 24px',
    width: style.width || '100%',
    height: style.height || '100%'
  }
})

const iconStyle = computed(() => {
  const style = props.control.styles?.icon || {}
  const defaultStyle = props.control.styles?.default || {}
  return {
    icon: style.icon || '',
    color: style.color || defaultStyle.color || '#fff',
    font_size: style.font_size || '18px'
  }
})

const handleClick = () => {
  emit('change', { type: 'click', control: props.control })
  if (props.control.events?.click?.key) {
    emit('action', { 
      type: 'click', 
      control: props.control,
      eventKey: 'click',
    })
  }
}
</script>

<style lang="scss" scoped>
.control-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  transition: all 0.3s;
  user-select: none;

  &:hover {
    opacity: 0.9;
  }

  &:active {
    transform: scale(0.98);
  }

  .button-icon {
    font-size: 18px;
  }

  .button-text {
    font-weight: 500;
  }
}
</style>
