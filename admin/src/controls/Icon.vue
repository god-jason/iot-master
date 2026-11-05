<template>
  <i 
    v-if="iconClass" 
    class="iotfont" 
    :class="iconClass"
    :style="iconStyle"
  ></i>
  <span v-else-if="text" :style="iconStyle">{{ text }}</span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  name: {
    type: String,
    default: ''
  },
  size: {
    type: [String, Number],
    default: '16px'
  },
  color: {
    type: String,
    default: ''
  },
  text: {
    type: String,
    default: ''
  },
  class: {
    type: String,
    default: ''
  }
})

const iconClass = computed(() => {
  if (!props.name) return ''
  return `iot-${props.name}`
})

const iconStyle = computed(() => {
  const style = {}
  if (props.size) {
    style.fontSize = typeof props.size === 'number' ? `${props.size}px` : props.size
  }
  if (props.color) {
    style.color = props.color
  }
  return style
})
</script>

<style scoped>
.iotfont {
  display: inline-block;
  font-size: inherit;
  line-height: 1;
}
</style>