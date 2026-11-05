<template>
  <view 
    v-if="visible" 
    class="ds-backtop" 
    :class="{ 'ds-show': visible }"
    @click="handleClick"
  >
    <DsIcon name="up_arrow" :size="24" color="#fff" />
  </view>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import DsIcon from './DsIcon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  threshold: {
    type: Number,
    default: 400
  },
  duration: {
    type: Number,
    default: 300
  }
})

const visible = ref(false)
let scrollListener = null

const handleScroll = () => {
  const scrollTop = uni.getSystemInfoSync().windowHeight || 0
  const pages = getCurrentPages()
  const page = pages[pages.length - 1]
  const scrollEl = page.$el?.querySelector('.scroll-view')
  if (scrollEl) {
    visible.value = scrollEl.scrollTop > props.threshold
  } else {
    visible.value = false
  }
}

const handleClick = () => {
  uni.pageScrollTo({
    scrollTop: 0,
    duration: props.duration
  })
}

onMounted(() => {
  scrollListener = setInterval(handleScroll, 100)
})

onUnmounted(() => {
  if (scrollListener) {
    clearInterval(scrollListener)
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-backtop {
  position: fixed;
  right: $spacing-lg;
  bottom: calc(#{$spacing-xl} + #{$safe-bottom});
  width: 80rpx;
  height: 80rpx;
  background: rgba(91, 163, 232, 0.9);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transform: translateY(20rpx);
  transition: $transition-base;
  z-index: 100;

  &.ds-show {
    opacity: 1;
    transform: translateY(0);
  }

  &:active {
    background: $primary-color;
  }
}
</style>