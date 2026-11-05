<template>
  <div class="control-video">
    <video
      :src="control.url || value || ''"
      controls
      :style="videoStyle"
    ></video>
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

const videoStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    width: '100%',
    borderRadius: style.border_radius || '8px'
  }
})
</script>

<style lang="scss" scoped>
.control-video {
  width: 100%;

  video {
    display: block;
    width: 100%;
    border-radius: 8px;
  }
}
</style>
