<template>
  <view class="ds-swiper-wrapper">
    <swiper
      :indicator-dots="indicatorDots"
      :autoplay="autoplay"
      :interval="interval"
      :duration="duration"
      :circular="circular"
      :indicator-color="indicatorColor"
      :indicator-active-color="indicatorActiveColor"
      :current="current"
      @change="handleChange"
      class="ds-custom-swiper"
    >
      <template v-if="items && items.length">
        <swiper-item v-for="(item, index) in items" :key="index" @click="handleItemClick(item, index)">
          <view class="ds-swiper-item">
            <image :src="item.image" mode="aspectFill" class="ds-swiper-image" />
            <view v-if="item.text" class="ds-swiper-text">{{ item.text }}</view>
          </view>
        </swiper-item>
      </template>
      <slot></slot>
    </swiper>
    <view v-if="pagination && (!items || items.length > 0)" class="ds-swiper-pagination">
      <view 
        v-for="(_, index) in (items ? items.length : 3)" 
        :key="index" 
        class="ds-pagination-dot"
        :class="{ 'ds-active': current === index }"
      ></view>
    </view>
  </view>
</template>

<script setup>
import { ref, watch } from 'vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  items: {
    type: Array,
    default: null
  },
  current: {
    type: Number,
    default: 0
  },
  indicatorDots: {
    type: Boolean,
    default: false
  },
  autoplay: {
    type: Boolean,
    default: false
  },
  interval: {
    type: Number,
    default: 3000
  },
  duration: {
    type: Number,
    default: 500
  },
  circular: {
    type: Boolean,
    default: false
  },
  pagination: {
    type: Boolean,
    default: false
  },
  indicatorColor: {
    type: String,
    default: 'rgba(255, 255, 255, 0.5)'
  },
  indicatorActiveColor: {
    type: String,
    default: '#fff'
  }
})

const emit = defineEmits(['update:current', 'change', 'click'])

const activeIndex = ref(props.current)

watch(() => props.current, (val) => {
  activeIndex.value = val
})

const handleChange = (e) => {
  activeIndex.value = e.detail.current
  emit('update:current', activeIndex.value)
  emit('change', activeIndex.value)
}

const handleItemClick = (item, index) => {
  emit('click', { item, index })
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-swiper-wrapper {
  position: relative;
  width: 100%;

  .ds-custom-swiper {
    height: 400rpx;

    .ds-swiper-item {
      position: relative;
      width: 100%;
      height: 100%;

      .ds-swiper-image {
        width: 100%;
        height: 100%;
      }

      .ds-swiper-text {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        padding: $spacing-base;
        background: linear-gradient(transparent, rgba(0, 0, 0, 0.6));
        color: #fff;
        font-size: $font-size-base;
      }
    }
  }

  .ds-swiper-pagination {
    position: absolute;
    bottom: $spacing-base;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    gap: $spacing-xs;

    .ds-pagination-dot {
      width: 12rpx;
      height: 12rpx;
      background: rgba(255, 255, 255, 0.5);
      border-radius: 50%;

      &.ds-active {
        background: #fff;
        width: 24rpx;
        border-radius: 6rpx;
      }
    }
  }
}
</style>