<template>
  <view class="ds-grid" :class="{ 'ds-bordered': bordered }" :style="gridStyle">
    <view 
      v-for="(item, index) in items" 
      :key="index" 
      class="ds-grid-item"
      :style="getItemStyle(item)"
      @click="handleItemClick(item, index)"
    >
      <view v-if="item.icon" class="ds-grid-icon">
        <DsIcon :name="item.icon" :size="iconSize" :color="item.color || '#5BA3E8'" />
      </view>
      <view v-if="item.image" class="ds-grid-icon">
        <image :src="item.image" mode="aspectFill" class="ds-grid-image" />
      </view>
      <text v-if="item.text" class="ds-grid-text">{{ item.text }}</text>
      <text v-if="item.badge" class="ds-grid-badge">{{ item.badge }}</text>
      <slot :name="`item-${index}`" :item="item"></slot>
    </view>
    <slot></slot>
  </view>
</template>

<script setup>
import { computed } from 'vue'
import DsIcon from './DsIcon.vue'

defineOptions({
  options: {
    virtualHost: true
  }
})

const props = defineProps({
  items: {
    type: Array,
    default: () => []
  },
  columns: {
    type: Number,
    default: 4
  },
  column: {
    type: Number,
    default: 0
  },
  gap: {
    type: Number,
    default: 16
  },
  gutter: {
    type: Number,
    default: 0
  },
  iconSize: {
    type: Number,
    default: 48
  },
  bordered: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['click'])

const resolvedColumns = computed(() => props.column || props.columns)

const resolvedGap = computed(() => props.gutter || props.gap)

const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${resolvedColumns.value}, 1fr)`,
  gap: `${resolvedGap.value}rpx`
}))

const getItemStyle = (item) => {
  if (item.span) {
    return { gridColumn: `span ${item.span}` }
  }
  return {}
}

const handleItemClick = (item, index) => {
  emit('click', { item, index })
}
</script>

<style lang="scss">
@import './variables.scss';

.ds-grid {
  display: grid;
  padding: $spacing-base;
  background: $bg-primary;
  border-radius: $radius-lg;

  &.ds-bordered {
    border: 2rpx solid $border-color;
  }

  .ds-grid-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: $spacing-lg $spacing-sm;
    position: relative;

    .ds-grid-icon {
      width: 100rpx;
      height: 100rpx;
      border-radius: $radius-lg;
      background: $bg-secondary;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-bottom: $spacing-sm;

      .ds-grid-image {
        width: 100%;
        height: 100%;
        border-radius: $radius-lg;
      }
    }

    .ds-grid-text {
      font-size: $font-size-sm;
      color: $text-primary;
      text-align: center;
    }

    .ds-grid-badge {
      position: absolute;
      top: 8rpx;
      right: 16rpx;
      min-width: 32rpx;
      height: 32rpx;
      padding: 0 8rpx;
      background: $danger-color;
      color: #fff;
      font-size: $font-size-xs;
      border-radius: 16rpx;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>