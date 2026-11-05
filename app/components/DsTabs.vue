<template>
  <view class="ds-tabs" :class="[`ds-tabs-${type}`]">
    <scroll-view v-if="scrollable" scroll-x class="ds-tabs-scroll" :show-scrollbar="false">
      <view class="ds-tabs-inner">
        <slot></slot>
      </view>
    </scroll-view>
    <view v-else class="ds-tabs-flex">
      <slot></slot>
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
  modelValue: {
    type: Number,
    default: 0
  },
  type: {
    type: String,
    default: 'line',
    validator: (val) => ['line', 'card'].includes(val)
  },
  scrollable: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const currentIndex = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  currentIndex.value = val
})

const handleTabClick = (index) => {
  if (currentIndex.value === index) return
  currentIndex.value = index
  emit('update:modelValue', index)
  emit('change', index)
}

defineExpose({
  currentIndex,
  handleTabClick
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-tabs {
  background: $bg-primary;

  &.ds-tabs-scroll {
    white-space: nowrap;
    padding: $spacing-sm 0;

    .ds-tabs-inner {
      display: inline-flex;
      align-items: center;
    }
  }

  &.ds-tabs-flex {
    display: flex;

    .ds-tab-item {
      flex: 1;
      justify-content: center;
      padding: $spacing-base 0;
    }
  }

  &.ds-tabs-card {
    padding: $spacing-sm;
    border-radius: $radius-lg;
    background: $bg-secondary;

    .ds-tab-item {
      padding: $spacing-sm $spacing-lg;
      border-radius: $radius-base;
      background: $bg-primary;
      margin: 0 $spacing-xs;

      &.ds-active {
        background: $primary-color;

        .ds-tab-text {
          color: #fff;
        }
      }
    }
  }
}
</style>