<template>
  <view
    class="ds-card"
    :class="[customClass, attrs.class, `ds-card-${size}`, { 'ds-shadow': shadow, 'ds-bordered': bordered }]"
    :style="[customStyle, attrs.style, cardStyle]"
    @click="handleClick"
  >
    <view v-if="title" class="ds-card-header">
      <text class="ds-card-title">{{ title }}</text>
      <view class="ds-card-header-extra">
        <slot name="header-extra"></slot>
      </view>
    </view>
    <view class="ds-card-body">
      <view class="ds-card-content">
        <slot></slot>
      </view>
    </view>
    <view v-if="$slots.footer" class="ds-card-footer">
      <slot name="footer"></slot>
    </view>
  </view>
</template>

<script setup>
import { computed, useAttrs } from 'vue'

defineOptions({
  inheritAttrs: false,
  options: {
    styleIsolation: 'apply-shared',
    virtualHost: true
  }
})

const attrs = useAttrs()

const props = defineProps({
  title: {
    type: String,
    default: ''
  },
  size: {
    type: String,
    default: 'default',
    validator: (val) => ['small', 'default', 'large'].includes(val)
  },
  shadow: {
    type: Boolean,
    default: true
  },
  bordered: {
    type: Boolean,
    default: false
  },
  padding: {
    type: String,
    default: ''
  },
  customClass: {
    type: String,
    default: ''
  },
  customStyle: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['click'])

const cardStyle = computed(() => {
  const style = {}
  if (props.padding) style.padding = props.padding
  return style
})

const handleClick = (event) => {
  emit('click', event)
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-card {
  // 用 background-color 而非 background 简写：
  // customClass 中的 background 简写（含渐变 image）无论源顺序如何都能覆盖此默认色
  background-color: $bg-primary;
  border-radius: $radius-lg;
  margin: $spacing-base $spacing-base 0;
  padding: $spacing-lg;

  &.ds-card-small {
    .ds-card-body {
      padding: $spacing-sm;
    }
  }

  &.ds-card-large {
    .ds-card-body {
      padding: $spacing-xl;
    }
  }

  &.ds-shadow {
    box-shadow: $shadow-sm;
  }

  &.ds-bordered {
    border: 2rpx solid $border-color;
  }

  .ds-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: $spacing-base $spacing-lg;
    border-bottom: 1rpx solid $border-light;

    .ds-card-title {
      font-size: $font-size-lg;
      font-weight: 600;
      color: $text-primary;
    }

    .ds-card-header-extra {
      font-size: $font-size-sm;
      color: $text-secondary;
    }
  }

  .ds-card-body {
    padding: 0;
    padding-top: $spacing-sm;
  }

  .ds-card-content {
    min-width: 0;
  }

  .ds-card-footer {
    padding: $spacing-base $spacing-lg;
    border-top: 1rpx solid $border-light;
    text-align: right;
  }
}
</style>
