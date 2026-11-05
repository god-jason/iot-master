<template>
  <view class="ds-collapse-item">
    <view class="ds-collapse-header" @click="handleClick">
      <text class="ds-collapse-title">{{ title }}</text>
      <DsIcon name="down_arrow" :size="20" color="#999" :class="{ 'ds-rotate': isActive }" />
    </view>
    <view class="ds-collapse-content" :class="{ 'ds-show': isActive }">
      <view class="ds-collapse-content-inner">
        <slot></slot>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, inject } from 'vue'
import DsIcon from './DsIcon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  title: {
    type: String,
    default: ''
  },
  index: {
    type: Number,
    default: 0
  }
})

const collapseContext = inject('collapseContext')

const isActive = computed(() => {
  if (!collapseContext) return false
  return collapseContext.activeNames.value.includes(props.index)
})

const handleClick = () => {
  collapseContext?.toggle(props.index)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-collapse-item {
  border-bottom: 1rpx solid $border-color;

  &:last-child {
    border-bottom: none;
  }

  .ds-collapse-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: $spacing-base;
    background: $bg-primary;

    .ds-collapse-title {
      font-size: $font-size-base;
      color: $text-primary;
    }

    .ds-rotate {
      transform: rotate(180deg);
    }
  }

  .ds-collapse-content {
    height: 0;
    overflow: hidden;
    transition: height 0.3s ease;
    background: $bg-secondary;

    &.ds-show {
      height: auto;
    }

    .ds-collapse-content-inner {
      padding: $spacing-base;
    }
  }
}
</style>