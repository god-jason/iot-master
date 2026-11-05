<template>
  <view class="control-video">
    <video
      class="video-player"
      :src="videoSrc"
      controls
    ></video>
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

const videoSrc = computed(() => {
  return props.control.url || props.value || ''
})
</script>

<style lang="scss" scoped>
.control-video {
  padding: 24rpx;

  .video-player {
    width: 100%;
    border-radius: 16rpx;
  }
}
</style>
