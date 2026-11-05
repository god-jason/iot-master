<template>
  <div class="control-image">
    <template v-if="hasImages">
      <el-carousel
        v-if="images.length > 1"
        :height="carouselHeight"
        indicator-position="outside"
        :style="carouselStyle"
      >
        <el-carousel-item v-for="(img, index) in images" :key="index">
          <img :src="img" class="control-image-item" :style="imageStyle" />
        </el-carousel-item>
      </el-carousel>
      <img v-else :src="images[0]" class="control-image-item" :style="imageStyle" />
    </template>
    <div v-else class="control-image-placeholder">
      <Icon name="camera" class="placeholder-icon" />
      <span class="placeholder-text">暂无图片</span>
    </div>
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

const images = computed(() => {
  const imgs = props.control.images
  return Array.isArray(imgs) ? imgs.filter(Boolean) : []
})

const hasImages = computed(() => images.value.length > 0)

const carouselHeight = computed(() => '160px')

const carouselStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    borderRadius: style.border_radius || '8px',
    overflow: 'hidden'
  }
})

const imageStyle = computed(() => {
  const style = props.control.styles?.default || {}
  return {
    width: '100%',
    borderRadius: style.border_radius || '8px',
    objectFit: 'cover'
  }
})
</script>

<style lang="scss" scoped>
.control-image {
  width: 100%;

  .control-image-item {
    display: block;
    width: 100%;
    border-radius: 8px;
    object-fit: cover;
  }

  .control-image-placeholder {
    width: 100%;
    height: 160px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: #f5f5f5;
    border-radius: 8px;
    color: #999;

    .placeholder-icon {
      font-size: 36px;
      margin-bottom: 8px;
    }

    .placeholder-text {
      font-size: 14px;
    }
  }
}
</style>
