<template>
  <view 
    class="control-button" 
    :style="buttonStyle"
    @tap="handleClick"
  >
    <Icon 
      v-if="iconStyle.icon" 
      :name="iconStyle.icon" 
      :color="iconStyle.color"
      :size="iconStyle.font_size"
      class="button-icon" 
    />
    <text class="button-text">{{ control.name }}</text>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import Icon from './Icon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['change', 'action'])

const buttonStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    backgroundColor: style.background_color || '#5BA3E8',
    color: style.color || '#fff',
    fontSize: style.font_size || '32rpx',
    borderRadius: style.border_radius || '16rpx',
    border: style.border || 'none',
    padding: style.padding || '32rpx 48rpx',
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
    font_size: style.font_size || '36rpx'
  }
})

const handleClick = () => {
  if (props.control.events?.click?.key) {
    emit('action', {
      type: 'click',
      controlId: props.control.id,
      eventKey: 'click'
    })
  }
  emit('change', { type: 'click', controlId: props.control.id })
}
</script>

<style lang="scss" scoped>
.control-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16rpx;
  transition: opacity 0.3s;
  box-sizing: border-box;
  min-height: 80rpx;
  user-select: none;
  -webkit-tap-highlight-color: transparent;

  &:active {
    opacity: 0.8;
  }

  .button-icon {
    font-size: 36rpx;
  }

  .button-text {
    font-weight: 500;
  }
}
</style>
