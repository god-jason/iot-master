<template>
  <view class="ds-image-wrapper" @click="handleClick">
    <image
      :src="src"
      :mode="mode"
      :lazy-load="lazyLoad"
      class="ds-custom-image"
      :class="[`ds-image-${mode}`]"
      @load="handleLoad"
      @error="handleError"
    />
    <view v-if="loading" class="ds-image-loading">
      <DsLoading type="dots" text="" />
    </view>
    <view v-if="error" class="ds-image-error">
      <DsIcon name="picture" size="48" color="#ccc" />
    </view>
  </view>
</template>

<script setup>
import { ref } from 'vue'
import DsIcon from './DsIcon.vue'
import DsLoading from './DsLoading.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  src: {
    type: String,
    default: ''
  },
  mode: {
    type: String,
    default: 'aspectFill'
  },
  lazyLoad: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['load', 'error', 'click'])

const loading = ref(true)
const error = ref(false)

const handleLoad = () => {
  loading.value = false
  error.value = false
  emit('load')
}

const handleError = () => {
  loading.value = false
  error.value = true
  emit('error')
}

const handleClick = () => {
  if (!error.value && props.src) {
    uni.previewImage({
      urls: [props.src],
      current: props.src
    })
    emit('click')
  }
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-image-wrapper {
  position: relative;
  overflow: hidden;
  background: $bg-secondary;

  .ds-custom-image {
    width: 100%;
    height: 100%;
    display: block;
  }

  .ds-image-loading, .ds-image-error {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .ds-image-error {
    width: 120rpx;
    height: 120rpx;
    background: $bg-primary;
    border-radius: $radius-base;
  }
}
</style>