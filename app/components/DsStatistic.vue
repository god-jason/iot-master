<template>
  <view class="ds-statistic">
    <view class="ds-statistic-icon" v-if="icon">
      <DsIcon :name="icon" :size="32" :color="iconColor" />
    </view>
    <view class="ds-statistic-content">
      <text class="ds-statistic-value" :style="{ color: valueColor }">{{ prefix }}{{ displayValue }}{{ suffix }}</text>
      <text class="ds-statistic-label">{{ label }}</text>
    </view>
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
  value: {
    type: [Number, String],
    default: 0
  },
  label: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: ''
  },
  iconColor: {
    type: String,
    default: '#5BA3E8'
  },
  valueColor: {
    type: String,
    default: '#333'
  },
  prefix: {
    type: String,
    default: ''
  },
  suffix: {
    type: String,
    default: ''
  },
  precision: {
    type: Number,
    default: 0
  }
})

const displayValue = computed(() => {
  const val = Number(props.value)
  return val.toFixed(props.precision)
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-statistic {
  display: flex;
  align-items: center;
  gap: $spacing-base;
  padding: $spacing-base;

  .ds-statistic-icon {
    width: 80rpx;
    height: 80rpx;
    border-radius: $radius-lg;
    background: $bg-secondary;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .ds-statistic-content {
    flex: 1;

    .ds-statistic-value {
      font-size: $font-size-xl;
      font-weight: 700;
      color: $text-primary;
      display: block;
      margin-bottom: $spacing-xs;
    }

    .ds-statistic-label {
      font-size: $font-size-sm;
      color: $text-placeholder;
    }
  }
}
</style>