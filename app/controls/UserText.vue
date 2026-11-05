<template>
  <view class="control-text" :style="textStyle">
    <text class="text-content">{{ control.text || value }}</text>
  </view>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  control: {
    type: Object,
    required: true
  },
  value: {
    type: [String, Number, Boolean],
    default: ''
  }
})

const textStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    color: style.color || '#333',
    fontSize: style.font_size || '28rpx',
    fontWeight: style.font_weight || '400',
    textAlign: style.text_align || 'left'
  }
})
</script>

<style lang="scss" scoped>
.control-text {
  padding: 24rpx;

  .text-content {
    display: block;
    line-height: 1.6;
  }
}
</style>
