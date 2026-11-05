<template>
  <view class="ds-no-more">
    <text>{{ text }}</text>
  </view>
</template>

<script setup>
defineOptions({
  options: {
    virtualHost: true
  }
})

defineProps({
  text: {
    type: String,
    default: '没有更多了'
  }
})
</script>

<style lang="scss">
@import './variables.scss';

.ds-no-more {
  padding: $spacing-lg 0;
  text-align: center;
  font-size: $font-size-sm;
  color: $text-placeholder;
}
</style>