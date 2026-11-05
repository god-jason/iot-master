<template>
  <text 
    class="iconfont" 
    :class="[iconClass, customClass]" 
    :style="iconStyle"
    @click="handleClick"
  ></text>
</template>

<script setup>
import { computed } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  name: {
    type: String,
    required: true
  },
  size: {
    type: [Number, String],
    default: 24
  },
  color: {
    type: String,
    default: '#333'
  },
  customClass: {
    type: String,
    default: ''
  },
  customStyle: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['click'])

const iconClass = computed(() => {
  let name = props.name
  if (name.startsWith('icon-')) {
    return name
  }
  return `icon-${name}`
})

const iconStyle = computed(() => {
  const baseStyle = {
    fontSize: typeof props.size === 'number' ? `${props.size}px` : props.size,
    color: props.color
  }
  return { ...baseStyle, ...props.customStyle }
})

const handleClick = (e) => {
  emit('click', e)
}
</script>
