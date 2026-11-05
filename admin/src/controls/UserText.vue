<template>
  <div class="control-text" :style="textStyle">
    {{ control.text || value }}
  </div>
</template>

<script setup>
import { computed } from 'vue'

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

const textStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    color: style.color || '#333',
    fontSize: style.font_size || '14px',
    fontWeight: style.font_weight || 'normal',
    textAlign: style.text_align || 'left'
  }
})
</script>

<style lang="scss" scoped>
.control-text {
  width: 100%;
  word-break: break-word;
  line-height: 1.5;
}
</style>
