<template>
  <view class="control-image">
    <swiper
      v-if="images.length > 1"
      class="image-swiper"
      :indicator-dots="true"
      :autoplay="true"
      :circular="true"
    >
      <swiper-item v-for="(img, index) in images" :key="index">
        <image
          class="swiper-image"
          :src="img"
          mode="widthFix"
        ></image>
      </swiper-item>
    </swiper>
    <image
      v-else-if="images.length === 1"
      class="single-image"
      :src="images[0]"
      mode="widthFix"
    ></image>
    <view v-else class="image-placeholder">
      <text class="placeholder-text">暂无图片</text>
    </view>
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

const images = computed(() => {
  const imgs = props.control.images
  if (Array.isArray(imgs)) {
    return imgs.filter((item) => !!item)
  }
  return []
})
</script>

<style lang="scss" scoped>
.control-image {
  padding: 24rpx;

  .image-swiper {
    width: 100%;
    height: 320rpx;

    .swiper-image {
      width: 100%;
      border-radius: 16rpx;
    }
  }

  .single-image {
    width: 100%;
    border-radius: 16rpx;
  }

  .image-placeholder {
    width: 100%;
    height: 240rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #f5f5f5;
    border-radius: 16rpx;

    .placeholder-text {
      font-size: 28rpx;
      color: #999;
    }
  }
}
</style>
